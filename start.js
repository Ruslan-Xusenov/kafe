import { spawn } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Color escape codes
const colors = {
  reset: "\x1b[0m",
  green: "\x1b[32m",
  blue: "\x1b[34m",
  magenta: "\x1b[35m",
  cyan: "\x1b[36m",
  red: "\x1b[31m",
  yellow: "\x1b[33m",
};

function runProcess(name, color, command, args, cwd) {
  console.log(`${color}[${name}] starting...${colors.reset}`);

  const isWindows = process.platform === "win32";
  const shell = isWindows ? true : false;

  const child = spawn(command, args, {
    cwd: path.resolve(__dirname, cwd),
    shell: shell,
    stdio: 'inherit',
    env: { ...process.env, FORCE_COLOR: 'true' }
  });

  child.on('error', (err) => {
    console.error(`${colors.red}[${name}] error: ${err.message}${colors.reset}`);
  });

  child.on('close', (code) => {
    console.log(`${colors.yellow}[${name}] process exited with code ${code}${colors.reset}`);
  });

  return child;
}

async function start() {
  console.log(`${colors.cyan}🚀 Kafe Automation System starting...${colors.reset}\n`);

  // 1. Start Backend API
  const backend = runProcess(
    "Backend API",
    colors.blue,
    "go",
    ["run", "./cmd/api/main.go"],
    "backend"
  );

  // 2. Start Telegram Bot
  const bot = runProcess(
    "Telegram Bot",
    colors.magenta,
    "go",
    ["run", "."],
    "telegram_bot"
  );

  // 3. Start Frontend (npm install if needed, but dev for now)
  const frontend = runProcess(
    "Frontend",
    colors.green,
    process.platform === "win32" ? "npm.cmd" : "npm",
    ["run", "dev"],
    "frontend"
  );

  // Handle termination
  const cleanup = () => {
    console.log(`\n${colors.cyan}🛑 Stopping all services...${colors.reset}`);
    backend.kill();
    bot.kill();
    frontend.kill();
    process.exit();
  };

  process.on('SIGINT', cleanup);
  process.on('SIGTERM', cleanup);
}

start();