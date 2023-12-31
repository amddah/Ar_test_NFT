const button = document.querySelector('.menu__button');
const menu = document.querySelector('.menu__body');
const close = document.querySelector('.menu__header button');
const overlay = document.querySelector('.menu__overlay');

function showMenu () {
	button.setAttribute('hidden', '');
	menu.removeAttribute('hidden');
	overlay.removeAttribute('hidden');
};

function hideMenu () {
	menu.setAttribute('hidden', '');
	overlay.setAttribute('hidden', '');
	button.removeAttribute('hidden');
};

button.addEventListener('click', showMenu);
close.addEventListener('click', hideMenu);
overlay.addEventListener('click', hideMenu);