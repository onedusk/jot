// Jot Documentation Search
(function() {
  let searchIndex = null;
  let searchInput = null;
  let searchResults = null;
  let searchOverlay = null;

  // Initialize search when DOM is ready
  document.addEventListener('DOMContentLoaded', function() {
    initializeSearch();
  });

  function initializeSearch() {
    // Create search UI elements
    createSearchUI();
    
    // Load search index
    loadSearchIndex();
    
    // Set up keyboard shortcuts
    setupKeyboardShortcuts();
  }

  function createSearchUI() {
    // Create search button in navigation
    const navHeader = document.querySelector('.nav-header');
    if (navHeader) {
      const searchButton = document.createElement('button');
      searchButton.className = 'search-button';
      searchButton.innerHTML = ' Search (Ctrl+K)';
      navHeader.appendChild(searchButton);
    }
    
    // Create search overlay
    searchOverlay = document.createElement('div');
    searchOverlay.className = 'search-overlay';
    searchOverlay.innerHTML = `
      <div class="search-modal">
        <div class="search-header">
          <input type="text" class="search-input" placeholder="Search documentation..." autocomplete="off">
          <button class="search-close" onclick="closeSearch()"></button>
        </div>
        <div class="search-results"></div>
      </div>
    `;
    
    // Get references to elements
    searchInput = searchOverlay.querySelector('.search-input');
    searchResults = searchOverlay.querySelector('.search-results');

    // Set up search input handler
    searchInput.addEventListener('input', debounce(performSearch, 300));
    
    // Close on overlay click
    searchOverlay.addEventListener('click', function(e) {
      if (e.target === searchOverlay) {
        closeSearch();
      }
    });
  }

  function loadSearchIndex() {
    // Calculate relative path to search index
    const depth = window.location.pathname.split('/').filter(p => p && p.includes('.html')).length - 1;
    const prefix = depth > 0 ? '../'.repeat(depth) : '';
    const indexPath = prefix + 'assets/search-index.json';

    fetch(indexPath)
      .then(response => response.json())
      .then(data => {
        searchIndex = data;
        console.log('Search index loaded:', searchIndex.documents.length, 'documents');
      })
      .catch(error => {
        console.error('Failed to load search index:', error);
      });
  }

  function setupKeyboardShortcuts() {
    document.addEventListener('keydown', function(e) {
      // Ctrl+K or Cmd+K to open search
      if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
        e.preventDefault();
        openSearch();
      }
      // Escape to close search
      if (e.key === 'Escape' && searchOverlay.classList.contains('active')) {
        closeSearch();
      }
    });
  }

  function openSearch() {
    searchOverlay.classList.add('active');
    searchInput.focus();
    searchInput.select();
  }

  window.closeSearch = function() {
    searchOverlay.classList.remove('active');
    searchInput.value = '';
    searchResults.innerHTML = '';
  }

  function performSearch() {
    const query = searchInput.value.toLowerCase().trim();
    
    if (!query) {
      searchResults.innerHTML = '<div class="search-empty">Type to search...</div>';
      return;
    }

    if (!searchIndex) {
      searchResults.innerHTML = '<div class="search-empty">Search index not loaded</div>';
      return;
    }

    // Search through documents
    const results = [];
    
    searchIndex.documents.forEach(doc => {
      let score = 0;
      let highlights = [];

      // Check title
      if (doc.title.toLowerCase().includes(query)) {
        score += 10;
        highlights.push({type: 'title', text: doc.title});
      }

      // Check headings
      doc.headings.forEach(heading => {
        if (heading.toLowerCase().includes(query)) {
          score += 5;
          highlights.push({type: 'heading', text: heading});
        }
      });

      // Check keywords
      doc.keywords.forEach(keyword => {
        if (keyword.toLowerCase().includes(query)) {
          score += 3;
        }
      });

      // Check content
      const contentLower = doc.content.toLowerCase();
      if (contentLower.includes(query)) {
        score += 1;
        
        // Extract context around match
        const index = contentLower.indexOf(query);
        const start = Math.max(0, index - 50);
        const end = Math.min(doc.content.length, index + query.length + 50);
        const context = '...' + doc.content.substring(start, end) + '...';
        highlights.push({type: 'content', text: context});
      }

      if (score > 0) {
        results.push({
          doc: doc,
          score: score,
          highlights: highlights
        });
      }
    });

    // Sort by score
    results.sort((a, b) => b.score - a.score);

    // Display results
    displaySearchResults(results.slice(0, 10)); // Show top 10 results
  }

  function displaySearchResults(results) {
    if (results.length === 0) {
      searchResults.innerHTML = '<div class="search-empty">No results found</div>';
      return;
    }

    const html = results.map(result => {
      const highlights = result.highlights.slice(0, 2).map(h => {
        if (h.type === 'title') {
          return `<div class="search-result-title">${escapeHtml(h.text)}</div>`;
        } else if (h.type === 'heading') {
          return `<div class="search-result-heading"> ${escapeHtml(h.text)}</div>`;
        } else if (h.type === 'content') {
          return `<div class="search-result-content">${escapeHtml(h.text)}</div>`;
        }
      }).join('');

      // Calculate relative path to result
      const depth = window.location.pathname.split('/').filter(p => p && p.includes('.html')).length - 1;
      const prefix = depth > 0 ? '../'.repeat(depth) : '';
      const path = prefix + result.doc.path;

      return `
        <a href="${path}" class="search-result">
          ${highlights}
          <div class="search-result-path">${result.doc.path}</div>
        </a>
      `;
    }).join('');

    searchResults.innerHTML = html;
  }

  function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
      const later = () => {
        clearTimeout(timeout);
        func(...args);
      };
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
    };
  }

  // Add search styles
  const style = document.createElement('style');
  style.textContent = `
    .search-button {
      width: 100%;
      padding: 0.5rem;
      margin: 0.5rem 0;
      background: var(--color-sky-500, #0ea5e9);
      color: white;
      border: none;
      border-radius: 0.375rem;
      cursor: pointer;
      font-size: 0.875rem;
      font-weight: 500;
      transition: background 0.2s;
    }

    .search-button:hover {
      background: var(--color-sky-400, #38bdf8);
    }

    .search-overlay {
      position: fixed;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background: rgba(0, 0, 0, 0.5);
      display: none;
      z-index: 1000;
      align-items: flex-start;
      justify-content: center;
      padding-top: 10vh;
    }

    .search-overlay.active {
      display: flex;
    }

    .search-modal {
      background: var(--bg-primary, white);
      border-radius: 0.5rem;
      width: 90%;
      max-width: 600px;
      max-height: 70vh;
      display: flex;
      flex-direction: column;
      box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
    }

    .search-header {
      display: flex;
      padding: 1rem;
      border-bottom: 1px solid var(--border-color, #e5e7eb);
    }

    .search-input {
      flex: 1;
      padding: 0.5rem;
      border: 1px solid var(--border-color, #e5e7eb);
      border-radius: 0.375rem;
      font-size: 1rem;
      outline: none;
      background: var(--bg-primary, white);
      color: var(--text-primary, #333);
    }

    .search-input:focus {
      border-color: var(--color-sky-500, #0ea5e9);
    }

    .search-close {
      margin-left: 0.5rem;
      padding: 0.5rem 0.75rem;
      background: transparent;
      border: none;
      cursor: pointer;
      font-size: 1.5rem;
      color: var(--text-secondary, #6b7280);
    }

    .search-results {
      flex: 1;
      overflow-y: auto;
      padding: 0.5rem;
    }

    .search-result {
      display: block;
      padding: 0.75rem;
      margin: 0.25rem 0;
      background: var(--bg-secondary, #f8f9fa);
      border-radius: 0.375rem;
      text-decoration: none;
      color: var(--text-primary, #333);
      transition: background 0.2s;
    }

    .search-result:hover {
      background: var(--bg-hover, #e5e7eb);
    }

    .search-result-title {
      font-weight: 600;
      font-size: 1rem;
      margin-bottom: 0.25rem;
      color: var(--text-primary, #111827);
    }

    .search-result-heading {
      font-size: 0.875rem;
      color: var(--color-sky-600, #0284c7);
      margin-bottom: 0.25rem;
    }

    .search-result-content {
      font-size: 0.875rem;
      color: var(--text-secondary, #6b7280);
      margin-bottom: 0.25rem;
    }

    .search-result-path {
      font-size: 0.75rem;
      color: var(--text-muted, #9ca3af);
    }

    .search-empty {
      text-align: center;
      padding: 2rem;
      color: var(--text-muted, #9ca3af);
    }

    @media (prefers-color-scheme: dark) {
      .search-modal {
        background: var(--bg-primary, #111827);
      }
      
      .search-result {
        background: var(--bg-secondary, #1f2937);
      }
      
      .search-result:hover {
        background: var(--bg-hover, #374151);
      }
    }
  `;
  document.head.appendChild(style);
})();