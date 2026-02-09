/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ['./views/**/*.{templ,gohtml,html}', './internal/**/*.{templ,gohtml,html}'],
    theme: {
        extend: {},
    },
    plugins: [require('@tailwindcss/typography')],
};
