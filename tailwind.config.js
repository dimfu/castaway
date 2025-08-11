/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ["./views/**/*.{html,js,templ}"],
	theme: {
		extend: {
			colors: {
				"antique-white": 'var(--antique-white)',
			},
			keyframes: {
				ellipsis: {
					'0%': { content: '""' },
					'33%': { content: '"."' },
					'66%': { content: '".."' },
					'100%': { content: '"..."' },
				}
			},
			animation: {
				ellipsis: 'ellipsis 1s steps(3, end) infinite'
			}
		},
	}
};
