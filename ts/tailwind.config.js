/** @type {import('tailwindcss').Config} */
export default {
    content: ["../**/*.html", "../**/*.templ"],
    theme: {
        extend: {
            colors: {
                theme: {
                    cream: "var(--cream)",
                    pink: "var(--pink)",
                    teal: "var(--teal)",
                    gold: "var(--gold)",
                    blue: "var(--blue)",
                    lavender: "var(--lavender)",
                    peach: "var(--peach)",
                    purple: "var(--purple)",
                    grey: "var(--grey)",
                    background: "var(--background)",
                    foreground: "var(--foreground)",
                },
            },
        },
    },
    plugins: [],
}

