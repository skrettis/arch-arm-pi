use rocket::launch;
use rocket::fs::{FileServer, relative};

#[launch]
fn rocket() -> _ {
    rocket::build().mount("/", FileServer::from(relative!("static")))
}
