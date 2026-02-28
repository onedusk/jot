// Jot Documentation Search — Inline Dropdown
(function() {
  var searchIndex = window.__searchIndex || null;
  var dropdown = null;
  var input = null;
  var selectedIndex = -1;
  var results = [];

  function getPrefix() {
    return (typeof window.__jotPrefix === 'string') ? window.__jotPrefix : '';
  }

  document.addEventListener('DOMContentLoaded', init);

  function init() {
    input = document.getElementById('header-search-input');
    if (!input) return;

    // Create dropdown container
    dropdown = document.createElement('div');
    dropdown.className = 'search-dropdown';
    input.parentElement.appendChild(dropdown);

    // Event listeners
    input.addEventListener('input', debounce(onSearch, 200));
    input.addEventListener('focus', onFocus);
    input.addEventListener('keydown', onKeydown);

    // Close on outside click
    document.addEventListener('click', function(e) {
      if (!e.target.closest('.header-search')) {
        close();
      }
    });

    // Cmd/Ctrl+K to focus
    document.addEventListener('keydown', function(e) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        input.focus();
        input.select();
      }
    });
  }

  function onFocus() {
    if (input.value.trim() && results.length > 0) {
      open();
    }
  }

  function onSearch() {
    var query = input.value.toLowerCase().trim();
    selectedIndex = -1;

    if (!query) {
      close();
      return;
    }

    if (!searchIndex) {
      renderEmpty('Search index not available');
      open();
      return;
    }

    results = [];

    searchIndex.documents.forEach(function(doc) {
      var score = 0;
      var context = '';

      // Title match
      if (doc.title.toLowerCase().indexOf(query) !== -1) {
        score += 10;
      }

      // Heading match
      if (doc.headings) {
        doc.headings.forEach(function(heading) {
          if (heading.toLowerCase().indexOf(query) !== -1) {
            score += 5;
          }
        });
      }

      // Keyword match
      if (doc.keywords) {
        doc.keywords.forEach(function(keyword) {
          if (keyword.toLowerCase().indexOf(query) !== -1) {
            score += 3;
          }
        });
      }

      // Content match
      var contentLower = doc.content.toLowerCase();
      var idx = contentLower.indexOf(query);
      if (idx !== -1) {
        score += 1;
        var start = Math.max(0, idx - 40);
        var end = Math.min(doc.content.length, idx + query.length + 40);
        context = (start > 0 ? '...' : '') + doc.content.substring(start, end) + (end < doc.content.length ? '...' : '');
      }

      if (score > 0) {
        results.push({ doc: doc, score: score, context: context });
      }
    });

    results.sort(function(a, b) { return b.score - a.score; });
    results = results.slice(0, 8);

    if (results.length === 0) {
      renderEmpty('No results for "' + escapeHtml(input.value) + '"');
    } else {
      renderResults(query);
    }
    open();
  }

  function renderResults(query) {
    var prefix = getPrefix();

    var html = results.map(function(result, i) {
      var path = prefix + result.doc.path;
      var title = highlightMatch(escapeHtml(result.doc.title), query);
      var context = result.context ? '<div class="search-dropdown-context">' + highlightMatch(escapeHtml(result.context), query) + '</div>' : '';
      var filePath = '<div class="search-dropdown-path">' + escapeHtml(result.doc.path) + '</div>';

      return '<a href="' + path + '" class="search-dropdown-item' + (i === selectedIndex ? ' selected' : '') + '" data-index="' + i + '">' +
        '<div class="search-dropdown-title">' + title + '</div>' +
        context +
        filePath +
        '</a>';
    }).join('');

    dropdown.innerHTML = html;
  }

  function renderEmpty(message) {
    dropdown.innerHTML = '<div class="search-dropdown-empty">' + message + '</div>';
  }

  function highlightMatch(text, query) {
    if (!query) return text;
    var escaped = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    var regex = new RegExp('(' + escaped + ')', 'gi');
    return text.replace(regex, '<mark>$1</mark>');
  }

  function onKeydown(e) {
    if (!dropdown.classList.contains('active') || results.length === 0) {
      if (e.key === 'Escape') {
        input.blur();
      }
      return;
    }

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      selectedIndex = Math.min(selectedIndex + 1, results.length - 1);
      updateSelection();
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      selectedIndex = Math.max(selectedIndex - 1, -1);
      updateSelection();
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (selectedIndex >= 0) {
        var items = dropdown.querySelectorAll('.search-dropdown-item');
        if (items[selectedIndex]) {
          window.location.href = items[selectedIndex].getAttribute('href');
        }
      }
    } else if (e.key === 'Escape') {
      close();
      input.blur();
    }
  }

  function updateSelection() {
    var items = dropdown.querySelectorAll('.search-dropdown-item');
    items.forEach(function(item, i) {
      if (i === selectedIndex) {
        item.classList.add('selected');
        item.scrollIntoView({ block: 'nearest' });
      } else {
        item.classList.remove('selected');
      }
    });
  }

  function open() {
    dropdown.classList.add('active');
  }

  function close() {
    dropdown.classList.remove('active');
    selectedIndex = -1;
  }

  function escapeHtml(text) {
    var div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  function debounce(func, wait) {
    var timeout;
    return function() {
      var args = arguments;
      clearTimeout(timeout);
      timeout = setTimeout(function() { func.apply(null, args); }, wait);
    };
  }
})();
