import index from "./src/index.html";

const API_ORIGIN = process.env.BUN_PUBLIC_API_URL ?? 'http://localhost:3000';

Bun.serve({
    port: 8788,
    routes: {
        "/": index,
        "/upload": index,
        "/api/*": (req) => {
            const url = new URL(req.url);
            return fetch(`${API_ORIGIN}${url.pathname}${url.search}`, {
                method: req.method,
                headers: req.headers,
                body: req.body,
            });
        },
    },
    development: {
        hmr: true,
        console: true,
    },
});

console.log(`Server running at http://localhost:8788 (API → ${API_ORIGIN})`);
