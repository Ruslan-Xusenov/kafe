const { spawn, execSync } = require('child_process');
const os = require('os');
const path = require('path');
const fs = require('fs');

const isWindows = os.platform() === 'win32';

const backendEnvPath = path.join(__dirname, 'backend', '.env');
const backendEnvExamplePath = path.join(__dirname, 'backend', '.env.example');
if (!fs.existsSync(backendEnvPath)) {
  console.log('⚠️  backend/.env topilmadi, .env.example dan nusxa olinmoqda...');
  fs.copyFileSync(backendEnvExamplePath, backendEnvPath);
}

const envContent = fs.readFileSync(backendEnvPath, 'utf8');
const dbPassMatch = envContent.match(/DB_PASSWORD=(.*)/);
const dbNameMatch = envContent.match(/DB_NAME=(.*)/);
const dbPass = dbPassMatch ? dbPassMatch[1].trim() : 'kafe2026';
const dbName = dbNameMatch ? dbNameMatch[1].trim() : 'kafe_db';

if (!isWindows) {
  try {
    console.log('🐘 PostgreSQL holati tekshirilmoqda...');
    const isActive = execSync('systemctl is-active postgresql || true').toString().trim() === 'active';
    if (!isActive) {
      console.log('🚀 PostgreSQL o\'chiq ekan, ishga tushirilmoqda...');
      execSync('sudo systemctl start postgresql');
    }
    
    console.log('⚙️  Ma\'lumotlar bazasi sozlanyapti...');
    execSync(`sudo -u postgres psql -c "ALTER USER postgres PASSWORD '${dbPass}';"`);
    execSync(`sudo -u postgres psql -c "CREATE DATABASE ${dbName};" || true`);
  } catch (err) {
    console.log('⚠️  PostgreSQL avtomat sozlashda xatolik (lekin davom etamiz):', err.message);
  }
}

console.log('\x1b[36m%s\x1b[0m', `
  _  __      ______ ______   _____  _            _______ 
 | |/ /     |  ____|  ____| |  __ \\| |          |__   __|
 | ' / __ _ | |__  | |__    | |__) | |     __ _    | |   
 |  < / _\` ||  __| |  __|   |  ___/| |    / _\` |   | |   
 | . \\ (_| || |    | |____  | |    | |___| (_| |   | |   
 |_|\\_\\__,_||_|    |______| |_|    |______\\__,_|   |_|   
                                                         
=========================================================
🚀 KAFE VA YETKAZIB BERISH TIZIMI ISHGA TUSHIRILMOQDA...
Bu skript Frontend va Backendni bir vaqtda ishlatadi.
=========================================================
`);

const backendCmd = isWindows ? 'go.exe' : 'go';
const backendArgs = ['run', 'cmd/api/main.go'];
const backendDir = path.join(__dirname, 'backend');

const frontendCmd = isWindows ? 'npm.cmd' : 'npm';
const frontendArgs = ['run', 'dev'];
const frontendDir = path.join(__dirname, 'frontend');

console.log('📦 \x1b[1mGo Backend\x1b[0m serveri ishga tushirilmoqda...');
const backendProcess = spawn(backendCmd, backendArgs, { cwd: backendDir, stdio: 'pipe', env: process.env });

backendProcess.stdout.on('data', data => process.stdout.write(`\x1b[34m[BACKEND]\x1b[0m ${data}`));
backendProcess.stderr.on('data', data => {
  const msg = data.toString();
  if (msg.includes('Listening and serving') || msg.includes('GIN-debug') || msg.includes('Server starting')) {
    process.stdout.write(`\x1b[36m[BACKEND INFO]\x1b[0m ${data}`);
  } else {
    process.stderr.write(`\x1b[31m[BACKEND LOG]\x1b[0m ${data}`);
  }
});

console.log('🎨 \x1b[1mReact Frontend\x1b[0m serveri ishga tushirilmoqda...\n');

const printerArgs = ['run', 'cmd/printer/main.go'];
console.log('🖨️  \x1b[1mChek printeri\x1b[0m ishga tushirilmoqda...\n');

setTimeout(() => {
  const frontendProcess = spawn(frontendCmd, frontendArgs, { cwd: frontendDir, stdio: 'pipe' });
  const printerProcess = spawn(backendCmd, printerArgs, { cwd: backendDir, stdio: 'pipe', env: process.env });

  frontendProcess.stdout.on('data', data => process.stdout.write(`\x1b[32m[FRONTEND]\x1b[0m ${data}`));
  frontendProcess.stderr.on('data', data => process.stderr.write(`\x1b[31m[FRONTEND XATO]\x1b[0m ${data}`));

  printerProcess.stdout.on('data', data => process.stdout.write(`\x1b[35m[PRINTER]\x1b[0m ${data}`));
  printerProcess.stderr.on('data', data => process.stderr.write(`\x1b[35m[PRINTER LOG]\x1b[0m ${data}`));

  console.log('\x1b[32m%s\x1b[0m', `
=========================================================
✅ BARCHA TIZIMLAR MUVAFFAQIYATLI ISHGA TUSHDI!

🌐 Frontend: http://localhost:5173
⚙️  Backend API: http://localhost:8080
🖨️  Chek printeri ulangan va kutmoqda...
=========================================================
Loyihani to'xtatish uchun terminalda CTRL+C ni bosing.
`);

  const killProcesses = () => {
    console.log('\n🛑 Tizimlar to\'xtatilmoqda...');
    if (isWindows) {
      spawn('taskkill', ['/pid', backendProcess.pid, '/f', '/t']);
      spawn('taskkill', ['/pid', frontendProcess.pid, '/f', '/t']);
      spawn('taskkill', ['/pid', printerProcess.pid, '/f', '/t']);
    } else {
      backendProcess.kill('SIGINT');
      frontendProcess.kill('SIGINT');
      printerProcess.kill('SIGINT');
    }
    process.exit();
  };

  process.on('SIGINT', killProcesses);
  process.on('SIGTERM', killProcesses);

}, 2000);