package handler

import (
	"net/http"
	"fmt"

)

func LandingPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8" />
            <meta name="viewport" content="width=device-width, initial-scale=1.0" />
            <title>Credora Banking App</title>
            <style>
                body {
                    font-family: system-ui, sans-serif;
                    background: #f9fafb;
                    color: #111827;
                    display: flex;
                    flex-direction: column;
                    align-items: center;
                    justify-content: center;
                    height: 100vh;
                    margin: 0;
                    text-align: center;
                }
                h1 { font-size: 2rem; margin-bottom: 1rem; }
                p { margin-bottom: 2rem; color: #4b5563; }
                a {
                    display: inline-block;
                    margin: 0.5rem;
                    padding: 0.75rem 1.5rem;
                    border-radius: 0.5rem;
                    text-decoration: none;
                    color: white;
                    background-color: #2563eb;
                    transition: background-color 0.2s;
                }
                a:hover { background-color: #1d4ed8; }
            </style>
        </head>
        <body>
            <h1>Welcome to Credora API</h1>
            <p>Your financial simulation and management platform backend.</p>
            <div>
                <a href="/api/v1/documentation" target="_blank">View API Documentation</a>
            </div>
        </body>
        </html>
    `)
}