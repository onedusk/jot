// Package renderer provides functionality for converting markdown documents into HTML.
// It uses the blackfriday library for markdown processing and includes features
// like syntax highlighting, task lists, and template-based page rendering.
package renderer

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} | Documentation</title>

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
                <div style="display: flex; align-items: center; gap: var(--spacing-xl);">
                    <button class="menu-toggle" onclick="toggleSidebar()">
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                        </svg>
                    </button>
                    <div class="logo">Documentation</div>
                </div>
                <nav class="header-nav">
                    <a href="#" class="header-link">Docs</a>
                    <a href="#" class="header-link">API</a>
                    <a href="#" class="header-link">GitHub</a>
                </nav>
            </div>
        </header>

        <!-- Sidebar -->
        <aside class="sidebar" id="sidebar">
            <!-- Search -->
            <div class="search-container">
                <svg class="search-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                <input type="text" class="search-input" placeholder="Search documentation...">
            </div>

            <!-- Navigation -->
            <nav>
                {{.Navigation}}
            </nav>
        </aside>

        <!-- Main Content -->
        <main class="main">
            <div class="content">
                <!-- Breadcrumbs -->
                <nav class="breadcrumbs">
                    {{range $i, $item := .Breadcrumb}}
                        {{if $i}}<span class="breadcrumb-separator">/</span>{{end}}
                        <a href="{{$item.Path}}" class="breadcrumb-link">{{$item.Title}}</a>
                    {{end}}
                </nav>

                <!-- Article Content -->
                <article>
                    {{.Content}}
                </article>
            </div>
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
            // Add copy button to all code blocks
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

        // Search functionality
        const searchInput = document.querySelector('.search-input');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                const query = e.target.value.toLowerCase();
                // Implement search logic here
                console.log('Searching for:', query);
            });
        }

        // Sidebar Dropdown functionality
        document.querySelectorAll('.nav-section-title').forEach(title => {
            title.addEventListener('click', () => {
                const section = title.parentElement;
                section.classList.toggle('open');
            });
        });
    </script>
</body>
</html>`
