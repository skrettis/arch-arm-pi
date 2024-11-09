use std::env;

use rocket::launch;
use rocket::fs::FileServer;
use rocket::config::Config;

const PATH: &str = "$HOME/.config/arch-jlake-co/static";

#[launch]
fn rocket() -> _ {

    let port = env::var("PORT").unwrap_or("80".to_string());
    let port = port.parse::<u16>().unwrap();
    let mut config = Config::default();
    config.port = port;

    let home: String = env::var("HOME").unwrap();
    let path = PATH.replace("$HOME", &home);

    rocket::custom(config).mount("/", FileServer::from(path))
}
