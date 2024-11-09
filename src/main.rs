use std::env;
use rocket::launch;
use rocket::fs::FileServer;

const PATH: &str = "$HOME/.config/arch-jlake-co/static";

#[launch]
fn rocket() -> _ {
    let home: String = env::var("HOME").unwrap();
    let path = PATH.replace("$HOME", &home);
    rocket::build().mount("/", FileServer::from(path))
}
