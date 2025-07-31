import {
    CheckEnvContainsEmailData,
    GetCredential,
    RunScan,
    SaveCredentials, SaveToken
} from "../wailsjs/go/main/App";
import {EventsOn} from "../wailsjs/runtime";

const LOGIN_PAGE = 'login'
const MAIN_PAGE = 'main'
const CHANGE_TOKEN_PAGE = 'change-token'

const SCAN_INTERVAL_MS = /*24 * 60 * */ 60 * 1000;

let countdownTimer = null;
let scanTimeout = null;
let remainingTime = SCAN_INTERVAL_MS;
let countdownStartedAt = null;

function formatDuration(ms) {
    if (ms < 0) ms = 0;

    const sec = Math.floor(ms/1000) % 60;
    const min = Math.floor(ms/60000) % 60;
    const hr = Math.floor(ms/3600000) % 24;
    const day = Math.floor(ms/(24*3600000));

    return `${day}:${hr.toString().padStart(2,'0')}:${min.toString().padStart(2,'0')}:${sec.toString().padStart(2,'0')}`;
}

function startAutoScanCountdown() {
    clearInterval(countdownTimer);
    clearTimeout(scanTimeout);

    countdownStartedAt = Date.now();
    const targetTime = countdownStartedAt + remainingTime;

    countdownTimer = setInterval(() => {
        const now = Date.now();
        const remnant = targetTime - now;
        remainingTime = remnant;

        document.querySelector('.status__label').innerText = 'Автоматическое сканирование через:';
        document.querySelector('.status__value').innerText = formatDuration(remnant);

        if (remnant <= 0) {
            clearInterval(countdownTimer);
        }
    }, 1000);

    scanTimeout = setTimeout(async () => {
        clearInterval(countdownTimer);
        clearTimeout(scanTimeout);
        await doScan();
        remainingTime = SCAN_INTERVAL_MS;
        startAutoScanCountdown();
    }, remainingTime);
}

async function doScan() {
    document.querySelector('.status__label').innerText = 'Сканирование почты';
    document.querySelector('.status__value').innerText = '';
    await RunScan().catch(err => console.error(err));
}

function showPage(name) {
    document.querySelectorAll('.page').forEach(el => el.classList.add('hidden'));
    const target = document.getElementById(name);
    if (target) target.classList.remove('hidden');
}

const emailString = "EMAIL"

async function goToMain(sender) {
    let mail = document.querySelector('#email').value.trim();
    if (!mail || sender !== LOGIN_PAGE) {
        mail = await GetCredential(emailString).then(
            (res) => {
                return res;
            }
        ).catch(err => console.log(err));
    }

    document.querySelector('.change-email__mail-text').innerText = mail;

    showPage(MAIN_PAGE);
}

async function login() {
    const flag = await CheckEnvContainsEmailData().then(res => {
        return res
    }).catch(err => console.error(err));

    if (flag) {
        document.querySelector('.change-email__mail-text').innerText = await GetCredential(emailString).then(
            (res) => {
                return res;
            }
        ).catch(err => console.log(err));

        showPage(MAIN_PAGE);
    } else {
        showPage(LOGIN_PAGE);
    }
}

document.querySelector('.authorization__submit').addEventListener('click', async () => {
    const emailFieldWrapper = document.querySelector('.email-field');
    const emailInput = document.getElementById("email");
    const passwordInput = document.querySelector('#password');
    const password = document.getElementById("password").value.trim();

    emailFieldWrapper.classList.remove('error');

    const email = emailInput.value.trim();

    const success = await SaveCredentials(email, password).catch(err => console.error(err));

    if (success) {
        await goToMain(LOGIN_PAGE);
    } else {
        emailFieldWrapper.classList.add('error');
        passwordInput.classList.add('error');
        emailInput.focus();
    }
});

document.querySelector('#email').addEventListener('input', () => {
    document.querySelector('.email-field')?.classList.remove('error');
    document.querySelector('#password')?.classList.remove('error');
});

document.querySelector('#password').addEventListener('input', () => {
    document.querySelector('.email-field')?.classList.remove('error');
    document.querySelector('#password')?.classList.remove('error');
});

document.querySelector('.authorization__token-change-submit').addEventListener('click', async () => {
        const fieldWrapper = document.querySelector('.token-field');
        const input = document.getElementById('token');

        fieldWrapper.classList.remove('error');

        const token = input.value.trim();

        const statusInfoContainer = document.querySelector('.authorization__token-change-status-info');
        statusInfoContainer.innerText = 'Проверка токена';

        const baseText = statusInfoContainer.innerText.replace(/\.\.\.$/, '').trim();
        let dotCount = 1;

        const dotsAnimation = setInterval(
            () => {
                dotCount = (dotCount + 1) % 4;
                statusInfoContainer.innerText = baseText + '.'.repeat(dotCount);
            }, 500
        );

        const resp = await SaveToken(token)
            .then(res => res)
            .catch(err => {
                console.error(err);
                return false;
            });

        if (resp) {
            fieldWrapper.classList.remove('error');
            showPage(MAIN_PAGE);
        } else {
            fieldWrapper.classList.add('error');
            input.focus();
        }

        clearInterval(dotsAnimation);
        statusInfoContainer.innerText = '';
    });

document.getElementById('token').addEventListener('input', () => {
        document.querySelector('.token-field')?.classList.remove('error');
    });


document.querySelector('.change-email').addEventListener('click', () => {
    showPage(LOGIN_PAGE);
});

document.querySelector('.main__scan-btn').addEventListener('click', async () => {
    clearInterval(countdownTimer);
    clearTimeout(scanTimeout);
    await doScan();
    remainingTime = SCAN_INTERVAL_MS;
    startAutoScanCountdown();
});

document.querySelector('.top-controls__change-token').addEventListener('click', () => {
    showPage(CHANGE_TOKEN_PAGE)
})


document.querySelector('.page__cancel-btn').addEventListener('click', () => {
    showPage(MAIN_PAGE);
})

document.addEventListener('DOMContentLoaded', async () => {
    await login();

    EventsOn("log", (msg) => {
        const statusValue = document.querySelector('.status__value');
        const statusLabel = document.querySelector('.status__label');
        if (!statusValue) return;

        if (msg.startsWith('Писем прочитано:')) {
            statusLabel.innerText = msg;
            return;
        }

        if (msg.length > 90) {
            statusValue.innerText = msg.slice(0, 87) + '...';
            return;
        }

        statusValue.innerText = msg;
    });
});