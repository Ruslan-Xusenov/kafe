// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::process::{Child, Command};
use std::sync::{Arc, Mutex};
use tauri::{AppHandle, Manager};

/// Global backend process handle
static BACKEND_PROCESS: std::sync::OnceLock<Arc<Mutex<Option<Child>>>> =
    std::sync::OnceLock::new();

/// Start the Go backend process bundled alongside the app
fn start_backend(app: &AppHandle) {
    let process_arc = BACKEND_PROCESS.get_or_init(|| Arc::new(Mutex::new(None)));

    let resource_dir = app
        .path()
        .resource_dir()
        .expect("Failed to get resource dir");

    #[cfg(target_os = "windows")]
    let backend_name = "kafe-api.exe";
    #[cfg(not(target_os = "windows"))]
    let backend_name = "kafe-api";

    let backend_path = resource_dir.join("binaries").join(backend_name);
    let backend_dir = resource_dir.clone();

    println!("[Tauri] Starting backend: {:?}", backend_path);

    match Command::new(&backend_path)
        .current_dir(&backend_dir)
        .spawn()
    {
        Ok(child) => {
            println!("[Tauri] Backend started with PID: {}", child.id());
            let mut lock = process_arc.lock().unwrap();
            *lock = Some(child);
        }
        Err(e) => {
            eprintln!("[Tauri] Failed to start backend: {}", e);
        }
    }
}

/// Stop the Go backend process
fn stop_backend() {
    if let Some(arc) = BACKEND_PROCESS.get() {
        let mut lock = arc.lock().unwrap();
        if let Some(mut child) = lock.take() {
            println!("[Tauri] Stopping backend (PID: {})", child.id());
            let _ = child.kill();
            let _ = child.wait();
        }
    }
}

#[tauri::command]
fn get_app_version() -> String {
    env!("CARGO_PKG_VERSION").to_string()
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .invoke_handler(tauri::generate_handler![get_app_version])
        .setup(|app| {
            start_backend(app.handle());
            std::thread::sleep(std::time::Duration::from_millis(1500));
            Ok(())
        })
        .on_window_event(|_window, event| {
            if let tauri::WindowEvent::Destroyed = event {
                stop_backend();
            }
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
