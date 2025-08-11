/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ["./views/**/*.{html,js,templ}"],
	theme: {
		extend: {
			colors: {
				"antique-white": 'var(--antique-white)',
			},
		},
	}
};
