package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
)

type Node struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*Node
	RelPath  string // Add RelPath field to store relative path
}

type FileServer struct {
	root      string
	tree      *Node
	mutex     sync.RWMutex
	watcher   *fsnotify.Watcher
	clients   map[chan struct{}]struct{}
	clientsMu sync.Mutex
}

func NewFileServer(root string) (*FileServer, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fs := &FileServer{
		root:    absRoot,
		clients: make(map[chan struct{}]struct{}),
	}

	// Build initial tree
	tree, err := fs.buildTree(absRoot)
	if err != nil {
		return nil, err
	}
	fs.tree = tree

	// Start watching directory
	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	fs.watcher = watcher
	go fs.watchChanges()
	return fs, nil
}

func (fs *FileServer) watchChanges() {
	for {
		select {
		case event, ok := <-fs.watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				// Rebuild tree
				tree, err := fs.buildTree(fs.root)
				if err != nil {
					log.Printf("Error rebuilding tree: %v", err)
					continue
				}

				fs.mutex.Lock()
				fs.tree = tree
				fs.mutex.Unlock()

				// Notify all clients
				fs.clientsMu.Lock()
				for client := range fs.clients {
					select {
					case client <- struct{}{}:
					default:
					}
				}
				fs.clientsMu.Unlock()

				// If a new directory was created, start watching it
				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						fs.watcher.Add(event.Name)
					}
				}
			}
		case err, ok := <-fs.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func (fs *FileServer) buildTree(path string) (*Node, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Calculate relative path from root
	relPath, err := filepath.Rel(fs.root, path)
	if err != nil {
		return nil, err
	}
	if relPath == "." {
		relPath = ""
	}

	node := &Node{
		Name:    filepath.Base(path),
		Path:    path,
		RelPath: relPath,
		IsDir:   info.IsDir(),
	}

	if info.IsDir() {
		files, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			childPath := filepath.Join(path, file.Name())
			child, err := fs.buildTree(childPath)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, child)
		}

		sort.Slice(node.Children, func(i, j int) bool {
			if node.Children[i].IsDir != node.Children[j].IsDir {
				return node.Children[i].IsDir
			}
			return node.Children[i].Name < node.Children[j].Name
		})
	}

	return node, nil
}

func (fs *FileServer) subscribeToChanges() chan struct{} {
	updateChan := make(chan struct{}, 1)
	fs.clientsMu.Lock()
	fs.clients[updateChan] = struct{}{}
	fs.clientsMu.Unlock()
	return updateChan
}

func (fs *FileServer) unsubscribe(updateChan chan struct{}) {
	fs.clientsMu.Lock()
	delete(fs.clients, updateChan)
	fs.clientsMu.Unlock()
}

func generateHTML(node *Node) template.HTML {
	var html strings.Builder

	if node.IsDir {
		html.WriteString(fmt.Sprintf("<div class=\"folder\" data-path=\"%s\">\n", node.RelPath))
		html.WriteString(fmt.Sprintf("    📁 %s\n", node.Name))
		if len(node.Children) > 0 {
			html.WriteString("<ul>\n")
			for _, child := range node.Children {
				html.WriteString("<li>\n")
				html.WriteString(string(generateHTML(child)))
				html.WriteString("</li>\n")
			}
			html.WriteString("</ul>\n")
		}
		html.WriteString("</div>\n")
	} else {
		// Use RelPath for download link
		downloadPath := "/download/" + node.RelPath
		downloadPath = strings.TrimPrefix(downloadPath, "/") // Remove leading slash if present
		html.WriteString(fmt.Sprintf(
			"<div class=\"file\" data-path=\"%s\">📄 <a href=\"/%s\">%s</a></div>\n",
			node.RelPath, downloadPath, node.Name,
		))
	}

	return template.HTML(html.String())
}

func main() {
	root := "./static"
	fs, err := NewFileServer(root)
	if err != nil {
		log.Fatalf("Error initializing file server: %v", err)
	}
	defer fs.watcher.Close()

	r := gin.Default()

	// Serve static files
	r.Static("/static", "./static")

	// Main page
	r.GET("/", func(c *gin.Context) {
		fs.mutex.RLock()
		html := generateHTML(fs.tree)
		fs.mutex.RUnlock()

		c.HTML(200, "index.html", gin.H{
			"Tree": html,
		})
	})

	// SSE endpoint for updates
	r.GET("/updates", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		updateChan := fs.subscribeToChanges()
		defer fs.unsubscribe(updateChan)

		c.Stream(func(w io.Writer) bool {
			<-updateChan

			fs.mutex.RLock()
			html := generateHTML(fs.tree)
			fs.mutex.RUnlock()

			c.SSEvent("update", html)
			return true
		})
	})

	// File download endpoint
	r.GET("/download/*path", func(c *gin.Context) {
		filePath := filepath.Join(fs.root, c.Param("path"))
		c.File(filePath)
	})

	// Load HTML template
	r.LoadHTMLFiles("templates/index.html")

	r.Run(":8080")
}
