// Package renderer provides functionality for converting markdown documents into HTML.
// It uses the blackfriday library for markdown processing and includes features
// like syntax highlighting, task lists, and template-based page rendering.
package renderer

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} | {{.ProjectName}}</title>

    <!-- Modern Font Stack -->
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">

    <link rel="stylesheet" href="{{.RelativePrefix}}assets/style.css">
    <link rel="stylesheet" href="{{.RelativePrefix}}assets/syntax-highlighting.css">
    <script src="{{.RelativePrefix}}assets/highlight.js"></script>
</head>
<body>
    <div class="layout">
        <!-- Header -->
        <header class="header">
            <div class="header-content">
                <div class="header-left">
                    <button class="menu-toggle" onclick="toggleSidebar()">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                        </svg>
                    </button>
                    <a href="{{.RelativePrefix}}index.html" class="logo">
                        <span class="logo-icon"></span>
                        {{.ProjectName}}
                    </a>
                </div>

                <div class="header-center">
                    <div class="header-search">
                        <svg class="search-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                        </svg>
                        <input type="text" class="search-input" placeholder="Find something..." id="header-search-input">
                        <kbd class="search-kbd">&#8984;K</kbd>
                    </div>
                </div>

                <nav class="header-nav">
                    {{range .NavLinks}}
                    <a href="{{.Href}}" class="header-link">{{.Label}}</a>
                    {{end}}
                    <span class="header-avatar">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" width="20" height="20">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                        </svg>
                    </span>
                    <a href="#" class="header-signin">Sign in</a>
                </nav>
            </div>
        </header>

        <!-- Sidebar -->
        <aside class="sidebar" id="sidebar">
            <nav>
                {{.Navigation}}
            </nav>
        </aside>

        <!-- Main Content -->
        <main class="main">
            <div class="content">
                <!-- Article Content -->
                <article>
                    {{.Content}}
                </article>

                <!-- Feedback Widget -->
                <div class="feedback">
                    <span>Was this page helpful?</span>
                    <button class="feedback-btn" onclick="this.classList.add('selected')">Yes</button>
                    <button class="feedback-btn" onclick="this.classList.add('selected')">No</button>
                </div>

                <!-- Page Navigation -->
                <nav class="page-nav">
                    {{if .PrevPage}}
                    <a href="{{.PrevPage.Path}}" class="page-nav-link page-nav-prev">
                        <span class="page-nav-direction">&larr; Previous</span>
                        <span class="page-nav-title">{{.PrevPage.Title}}</span>
                    </a>
                    {{else}}
                    <span></span>
                    {{end}}
                    {{if .NextPage}}
                    <a href="{{.NextPage.Path}}" class="page-nav-link page-nav-next">
                        <span class="page-nav-direction">Next &rarr;</span>
                        <span class="page-nav-title">{{.NextPage.Title}}</span>
                    </a>
                    {{end}}
                </nav>
            </div>

            <!-- Footer -->
            <footer class="footer">
                <div class="footer-content">
                    <span class="footer-copyright">&copy; Copyright 2026. All rights reserved.</span>
                    <span class="footer-brand">{{.ProjectName}}</span>
                </div>
            </footer>
        </main>
    </div>

    <script>
        // Toggle sidebar on mobile
        function toggleSidebar() {
            const sidebar = document.getElementById('sidebar');
            sidebar.classList.toggle('open');
        }

        // Copy code functionality
        document.addEventListener('DOMContentLoaded', function() {
            document.querySelectorAll('pre').forEach(pre => {
                const button = document.createElement('button');
                button.className = 'copy-button';
                button.textContent = 'Copy';
                button.onclick = function() {
                    const code = pre.querySelector('code');
                    const text = code ? code.textContent : pre.textContent;

                    navigator.clipboard.writeText(text).then(() => {
                        button.textContent = 'Copied!';
                        button.classList.add('copied');

                        setTimeout(() => {
                            button.textContent = 'Copy';
                            button.classList.remove('copied');
                        }, 2000);
                    });
                };
                pre.appendChild(button);
            });
        });

    </script>
    <script>window.__jotPrefix = "{{.RelativePrefix}}";</script>
    <script src="{{.RelativePrefix}}assets/search-index.js"></script>
    <script src="{{.RelativePrefix}}assets/search.js"></script>
</body>
</html>`
