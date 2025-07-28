import {DisplayLogInfo, SendInfo, ValidateEmail} from "../wailsjs/go/main/App";

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

document.querySelector('.authorization__submit').addEventListener('click', async () => {
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
});

document.querySelector('.info').addEventListener('click', () => {
    showPage(LOGIN_PAGE);
});


document.addEventListener('DOMContentLoaded', () => {
    showPage(LOGIN_PAGE);
});