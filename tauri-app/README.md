# KafePlat Tauri Desktop App

Bu papka KafePlat ning Tauri asosidagi Windows desktop versiyasini o'z ichiga oladi.

## Tuzilishi

```
tauri-app/
├── package.json              — npm scripts
├── src-tauri/
│   ├── Cargo.toml            — Rust konfiguratsiya
│   ├── tauri.conf.json       — Tauri konfiguratsiya (oyna, bundle, installer)
│   ├── build.rs              — Build script
│   ├── .env                  — Backend .env (nusxa)
│   ├── icons/                — App ikonalar
│   │   ├── 32x32.png
│   │   ├── 128x128.png
│   │   ├── 128x128@2x.png
│   │   ├── icon.ico          — Windows uchun
│   │   └── icon.icns         — macOS uchun
│   ├── binaries/
│   │   └── kafe-api.exe      — Go backend (Windows uchun compile qilingan)
│   └── src/
│       └── main.rs           — Rust kodi (backend process boshqaradi)
```

## Build qilish

### GitHub Actions orqali (tavsiya)
1. GitHub'ga push qiling
2. `Actions` → `Build KafePlat Windows .exe` → `Run workflow`
3. Artifacts'dan `.exe` yuklab oling

### Lokal (agar Windows SDK bo'lsa)
```bash
. "$HOME/.cargo/env"
cd tauri-app
npm run build
```

## Qanday ishlaydi?

1. Foydalanuvchi `KafePlat-Setup.exe` o'rnatadi
2. App ochilganda Rust kodi `kafe-api.exe` ni ishga tushiradi (port 8081)
3. WebView2 (Windows'da o'rnatilgan) frontend UI ni ko'rsatadi
4. App yopilganda backend ham to'xtaydi

## Muhim eslatmalar

- `kafe-api.exe` — PostgreSQL'ga ulanadi, shuning uchun DB sozlanishi kerak
- `.env` fayli backend sozlamalari uchun app yonida joylashadi
- Windows'da `WebView2` kerak (Windows 10/11 da o'rnatilgan)
