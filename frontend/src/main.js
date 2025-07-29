import {CheckEnvContainsEmailData, GetCredential, RunScan, SaveCredentials} from "../wailsjs/go/main/App";
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

    document.querySelector('.info__mail-text').innerText = mail;

    showPage(MAIN_PAGE);
}

async function login() {
    const flag = await CheckEnvContainsEmailData().then(res => {
        return res
    }).catch(err => console.error(err));

    if (flag) {
        document.querySelector('.info__mail-text').innerText = await GetCredential(emailString).then(
            (res) => {
                return res;
            }
        ).catch(err => console.log(err));

        showPage(MAIN_PAGE);
    } else {
        showPage(LOGIN_PAGE);
    }
}

const success = "Успешно сохранено";

document.querySelector('.authorization__submit').addEventListener('click', async () => {
    const email = document.getElementById("email").value.trim();
    const password = document.getElementById("password").value.trim();

    const result = await SaveCredentials(email, password).catch(err => console.error(err));

    if (result === success) {
        await goToMain(LOGIN_PAGE)
    }
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

        if (msg.length > 90) {
            statusValue.innerText = msg.slice(0, 87) + '...';
            return;
        }

        if (msg.includes('...')) {
            const baseText = msg.replace(/\.\.\.$/, '').trim();
            statusValue.innerText = baseText;

            let dotCount = 0;
            dotsInterval = setInterval(() => {
                dotCount = (dotCount + 1) % 4;
                statusValue.innerText = baseText + '.'.repeat(dotCount);
            }, 500);
        } else {
            statusValue.innerText = msg;
        }
    });

});