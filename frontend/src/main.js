import {CheckEnvContainsEmailData, DisplayLogInfo, RunScan, SendInfo, ValidateEmail} from "../wailsjs/go/main/App";
import {EventsOn} from "../wailsjs/runtime";

const LOGIN_PAGE = 'login'
const MAIN_PAGE = 'main'

function showPage(name) {
    const pages = document.querySelectorAll('.page');
    pages.forEach((el) => el.classList.add('hidden'));

    const target = document.getElementById(name);
    if (target) {
        target.classList.remove('hidden');
    }

}

async function goToMain() {
    const email = document.querySelector('#email').value;
    const password = document.querySelector('#password').value;

    if (!email || !password) {
        DisplayLogInfo('Все поля должны быть заполнены').catch(err => console.error(err));

        return;
    }

    const emailValidated = await ValidateEmail(email).then(res => {
        if (!res) {
            DisplayLogInfo('Введен некорректный email').catch(err => console.error(err));
        }

        return res;
    }).catch(err => console.error(err));

    if (!emailValidated) {
        return;
    }

    SendInfo(email, password).catch(err => console.error(err));

    showPage(MAIN_PAGE);
}

async function login() {
    const flag = await CheckEnvContainsEmailData().then(res => {
        return res
    }).catch(err => console.error(err));

    if (flag) {
        showPage(MAIN_PAGE);
    } else {
        showPage(LOGIN_PAGE);
    }
}

document.querySelector('.authorization__submit').addEventListener('click', async () => {
    await goToMain()
});

document.querySelector('.info').addEventListener('click', () => {
    showPage(LOGIN_PAGE);
});

document.querySelector('.main__scan-btn').addEventListener('click', async () => {
    document.querySelector('.status__label').innerText = 'Идёт сканирование почты';

    await RunScan().catch(err => console.error(err));
});

let dotsInterval = null;

document.addEventListener('DOMContentLoaded', async () => {
    await login();

    EventsOn("log", (msg) => {
        const statusValue = document.querySelector('.status__value');
        if (!statusValue) return;

        clearInterval(dotsInterval);
        statusValue.innerText = msg;

        if (msg.includes('...')) {
            const baseText = msg.replace(/\.\.\.$/, '').trim();
            statusValue.innerText = baseText;

            let dotCount = 0;
            dotsInterval = setInterval(() => {
                dotCount = (dotCount + 1) % 4; // от 0 до 3
                statusValue.innerText = baseText + '.'.repeat(dotCount);
            }, 500);
        }
    });
});