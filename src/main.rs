use rocket::fs::FileServer;
use rocket::http::Method;
// use rocket::response::content::{self, RawHtml};
use rocket::route::{Handler, Outcome};
// use rocket::{get, launch, routes, Response, Route};
use rocket::{launch, Config, Response, Route};
use std::env;

const PATH: &str = "$HOME/.config/arch-jlake-co/static";

#[derive(Clone)]
struct CustomHandler {}

#[rocket::async_trait]
impl Handler for CustomHandler {
    async fn handle<'r>(&self, req: &'r rocket::Request<'_>, _: rocket::Data<'r>) -> Outcome<'r> {
        let mut resp: Response = Response::new();

        // make body a list of children in html of the folder structure
        // this would be say PATH/aarch64
        let home = env::var("HOME").unwrap();
        let path = PATH.replace("$HOME", &home);

        // use request uri to ensure path is proper
        let uri = req.uri().path();
        let path = format!("{}/{}", path, uri);

        let mut body = String::new();
        body.push_str("<html><body>");
        body.push_str("<h1>Arch Linux</h1>");
        body.push_str("<ul>");
        for entry in std::fs::read_dir(&path).unwrap() {
            let entry = entry.unwrap();
            let p = entry.path();
            let p = p.to_str().unwrap();
            let p = p.replace(path.as_str(), "");
            // let p = p.replace(".html", "");

            body.push_str(&format!("<li><a href=\"{}\">{}</a></li>", p, p));
        }

        // let body = format!("Path: {}", req.uri().path());

        resp.set_header(rocket::http::ContentType::HTML);

        resp.set_sized_body(
            body.clone().as_str().len(),
            std::io::Cursor::new(body.clone()),
        );

        Outcome::Success(resp)
    }
}

// recurse PATH to find all folders and make paths to them
// include all parents in child paths
impl Into<Vec<Route>> for CustomHandler {
    fn into(self) -> Vec<Route> {
        let home = env::var("HOME").unwrap();
        let path = PATH.replace("$HOME", &home);
        let mut routes = Vec::new();

        // ensures basic empty '/' is handled
        let route = Route::new(Method::Get, "/", self.clone());
        routes.push(route);

        // TODO: put this into a separate function for easy recursion
        // all it needs to do is return route paths
        for entry in std::fs::read_dir(&path).unwrap() {
            let entry = entry.unwrap();
            let p = entry.path();
            if p.is_dir() {
                let folder_name = p.file_name().unwrap().to_str().unwrap().to_string();
                let route_path = format!("/{}", folder_name);

                let route = Route::new(Method::Get, route_path.clone().as_str(), self.clone());

                routes.push(route);
            }
        }

        routes
    }
}

#[launch]
fn rocket() -> _ {
    let port = env::var("PORT").unwrap_or("80".to_string());
    let port = port.parse::<u16>().unwrap();
    let mut config = Config::default();
    config.port = port;
    config.address = "0.0.0.0".parse().unwrap();

    let home: String = env::var("HOME").unwrap();
    let path = PATH.replace("$HOME", &home);
    rocket::custom(config)
        .mount("/", CustomHandler {})
        .mount("/", FileServer::from(path))
}
// use std::env;

// use rocket::config::Config;
// use rocket::fs::FileServer;
// use rocket::response::content::{self, RawHtml};
// use rocket::{get, launch, routes};

// const PATH: &str = "$HOME/.config/arch-jlake-co/static";

// #[get("/")]
// fn root() -> RawHtml<String> {
//     let home: String = env::var("HOME").unwrap();
//     // let path = format!("{}/root.html", PATH.replace("$HOME", &home));
//     let path = PATH.replace("$HOME", &home);

//     // let contents = std::fs::read_to_string(path).unwrap();

//     let mut contents = String::new();
//     contents.push_str("<html><body>");
//     contents.push_str("<h1>Arch Linux</h1>");
//     contents.push_str("<ul>");
//     for entry in std::fs::read_dir(&path).unwrap() {
//         let entry = entry.unwrap();
//         let p = entry.path();
//         let p = p.to_str().unwrap();
//         let p = p.replace(path.as_str(), "");
//         let p = p.replace(".html", "");
//         // ignore "root"
//         if p == "/root" {
//             continue;
//         }
//         contents.push_str(&format!("<li><a href=\"{}\">{}</a></li>", p, p));
//     }
//     contents.push_str("</ul>");
//     contents.push_str("</body></html>");

//     content::RawHtml(contents)
// }

// #[launch]
// fn rocket() -> _ {
//     let port = env::var("PORT").unwrap_or("80".to_string());
//     let port = port.parse::<u16>().unwrap();
//     let mut config = Config::default();
//     config.port = port;
//     config.address = "0.0.0.0".parse().unwrap();

//     let home: String = env::var("HOME").unwrap();
//     let path = PATH.replace("$HOME", &home);

//     rocket::custom(config)
//         .mount("/", routes![root])
//         .mount("/", FileServer::from(path))
// }
