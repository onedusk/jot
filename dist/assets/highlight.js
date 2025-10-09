/**
 * Jot Documentation - Syntax Highlighting
 * A lightweight syntax highlighter for code blocks
 */

class SyntaxHighlighter {
    constructor() {
        this.patterns = {
            bash: [
                { pattern: /#.*$/gm, className: 'hljs-comment' },
                { pattern: /'([^'\\]|\\.)*'/g, className: 'hljs-string' },
                { pattern: /"([^"\\]|\\.)*"/g, className: 'hljs-string' },
                { pattern: /`([^`\\]|\\.)*`/g, className: 'hljs-string' },
                { pattern: /\b\d+\b/g, className: 'hljs-number' },
                { pattern: /\b(if|then|else|elif|fi|for|do|done|while|until|case|esac|function|return|export|local|readonly|declare)\b/g, className: 'hljs-keyword' },
                { pattern: /\b(echo|cat|ls|cd|pwd|mkdir|rmdir|rm|cp|mv|find|grep|sed|awk|sort|uniq|head|tail|less|more|wc|curl|wget|git|npm|yarn|docker|kubectl|jot|jotdoc)\b/g, className: 'hljs-built_in' }
            ],
            go: [
                { pattern: /\/\/.*$/gm, className: 'hljs-comment' },
                { pattern: /\/\*[\s\S]*?\*\//g, className: 'hljs-comment' },
                { pattern: /"([^"\\]|\\.)*"/g, className: 'hljs-string' },
                { pattern: /`([^`\\]|\\.)*`/g, className: 'hljs-string' },
                { pattern: /\b\d+(\.\d+)?\b/g, className: 'hljs-number' },
                { pattern: /\b(break|case|chan|const|continue|default|defer|else|fallthrough|for|func|go|goto|if|import|interface|map|package|range|return|select|struct|switch|type|var)\b/g, className: 'hljs-keyword' },
                { pattern: /\b(bool|byte|complex64|complex128|error|float32|float64|int|int8|int16|int32|int64|rune|string|uint|uint8|uint16|uint32|uint64|uintptr)\b/g, className: 'hljs-type' },
                { pattern: /\b(append|cap|close|complex|copy|delete|imag|len|make|new|panic|print|println|real|recover)\b/g, className: 'hljs-built_in' }
            ],
            javascript: [
                { pattern: /\/\/.*$/gm, className: 'hljs-comment' },
                { pattern: /\/\*[\s\S]*?\*\//g, className: 'hljs-comment' },
                { pattern: /'([^'\\]|\\.)*'/g, className: 'hljs-string' },
                { pattern: /"([^"\\]|\\.)*"/g, className: 'hljs-string' },
                { pattern: /`([^`\\]|\\.)*`/g, className: 'hljs-string' },
                { pattern: /\b\d+(\.\d+)?\b/g, className: 'hljs-number' },
                { pattern: /\b(async|await|break|case|catch|class|const|continue|debugger|default|delete|do|else|export|extends|finally|for|function|if|import|in|instanceof|let|new|return|super|switch|this|throw|try|typeof|var|void|while|with|yield)\b/g, className: 'hljs-keyword' },
                { pattern: /\b(Array|Boolean|Date|Error|Function|Number|Object|RegExp|String|console|document|window)\b/g, className: 'hljs-built_in' }
            ],
            python: [
                { pattern: /#.*$/gm, className: 'hljs-comment' },
                { pattern: /"""[\s\S]*?"""/g, className: 'hljs-string' },
                { pattern: /'''[\s\S]*?'''/g, className: 'hljs-string' },
                { pattern: /'([^'\\]|\\.)*'/g, className: 'hljs-string' },
                { pattern: /"([^"\\]|\\.)*"/g, className: 'hljs-string' },
                { pattern: /\b\d+(\.\d+)?\b/g, className: 'hljs-number' },
                { pattern: /\b(and|as|assert|break|class|continue|def|del|elif|else|except|exec|finally|for|from|global|if|import|in|is|lambda|not|or|pass|print|raise|return|try|while|with|yield)\b/g, className: 'hljs-keyword' },
                { pattern: /\b(bool|int|float|str|list|dict|tuple|set)\b/g, className: 'hljs-type' },
                { pattern: /\b(abs|all|any|bin|bool|chr|dir|enumerate|eval|filter|float|format|hex|id|input|int|isinstance|len|list|map|max|min|oct|open|ord|pow|print|range|repr|reversed|round|sorted|str|sum|type|zip)\b/g, className: 'hljs-built_in' }
            ],
            yaml: [
                { pattern: /#.*$/gm, className: 'hljs-comment' },
                { pattern: /'([^'\\]|\\.)*'/g, className: 'hljs-string' },
                { pattern: /"([^"\\]|\\.)*"/g, className: 'hljs-string' },
                { pattern: /\b\d+(\.\d+)?\b/g, className: 'hljs-number' },
                { pattern: /\b(true|false|null|yes|no|on|off)\b/g, className: 'hljs-keyword' },
                { pattern: /^(\s*)([a-zA-Z_][a-zA-Z0-9_]*)\s*:/gm, className: 'hljs-attr', replacement: '$1<span class="hljs-attr">$2</span>:' }
            ],
            json: [
                { pattern: /'([^'\\]|\\.)*'/g, className: 'hljs-string' },
                { pattern: /"([^"\\]|\\.)*"/g, className: 'hljs-string' },
                { pattern: /\b\d+(\.\d+)?\b/g, className: 'hljs-number' },
                { pattern: /\b(true|false|null)\b/g, className: 'hljs-keyword' },
                { pattern: /"([^"\\]|\\.)*"\s*:/g, className: 'hljs-attr' }
            ],
            rust: [
                { pattern: /\/\/.*$/gm, className: 'hljs-comment' },
                { pattern: /\/\*[\s\S]*?\*\//g, className: 'hljs-comment' },
                { pattern: /'([^'\\]|\\.)*'/g, className: 'hljs-string' },
                { pattern: /"([^"\\]|\\.)*"/g, className: 'hljs-string' },
                { pattern: /\b\d+(\.\d+)?\b/g, className: 'hljs-number' },
                { pattern: /\b(as|break|const|continue|crate|else|enum|extern|false|fn|for|if|impl|in|let|loop|match|mod|move|mut|pub|ref|return|self|Self|static|struct|super|trait|true|type|unsafe|use|where|while)\b/g, className: 'hljs-keyword' },
                { pattern: /\b(bool|char|str|i8|i16|i32|i64|u8|u16|u32|u64|f32|f64|isize|usize)\b/g, className: 'hljs-type' },
                { pattern: /\b(Some|None|Ok|Err|Vec|String|Option|Result)\b/g, className: 'hljs-built_in' }
            ],
            pseudocode: [
                { pattern: /\/\/.*$/gm, className: 'hljs-comment' },
                { pattern: /'([^'\\]|\\.)*'/g, className: 'hljs-string' },
                { pattern: /"([^"\\]|\\.)*"/g, className: 'hljs-string' },
                { pattern: /\b\d+(\.\d+)?\b/g, className: 'hljs-number' },
                { pattern: /\b(FUNCTION|IF|THEN|ELSE|ENDIF|FOR|WHILE|DO|RETURN|PROGRAM|BEGIN|END|SWITCH|CASE|DEFAULT|BREAK|CONTINUE|TRY|CATCH|THROW|EACH|IN|AND|OR|NOT)\b/g, className: 'hljs-keyword' },
                { pattern: /\b(TRUE|FALSE|NULL)\b/g, className: 'hljs-literal' }
            ]
        };
    }

    init() {
        if (document.readyState === 'loading') {
            document.addEventListener('DOMContentLoaded', () => this.highlightAll());
        } else {
            this.highlightAll();
        }
    }

    highlightAll() {
        const codeBlocks = document.querySelectorAll('pre code');
        codeBlocks.forEach(block => this.highlightBlock(block));
    }

    highlightBlock(block) {
        const language = this.detectLanguage(block);
        if (language && this.patterns[language]) {
            const originalText = block.textContent;
            const highlightedText = this.applyHighlighting(originalText, language);
            block.innerHTML = highlightedText;
        }
    }

    detectLanguage(block) {
        // Check block class
        const blockClasses = (block.className || '').split(' ');
        for (const cls of blockClasses) {
            if (cls.startsWith('language-')) {
                const lang = cls.replace('language-', '');
                if (this.patterns[lang]) return lang;
            }
        }

        // Check parent pre class
        const pre = block.parentElement;
        if (pre && pre.tagName === 'PRE') {
            const preClasses = (pre.className || '').split(' ');
            for (const cls of preClasses) {
                if (cls.startsWith('language-')) {
                    const lang = cls.replace('language-', '');
                    if (this.patterns[lang]) return lang;
                }
            }
        }

        return null;
    }

    applyHighlighting(text, language) {
        let result = this.escapeHtml(text);
        const patterns = this.patterns[language];

        // Apply patterns in order
        for (const pattern of patterns) {
            if (pattern.replacement) {
                result = result.replace(pattern.pattern, pattern.replacement);
            } else {
                result = result.replace(pattern.pattern, (match) => {
                    return `<span class="${pattern.className}">${match}</span>`;
                });
            }
        }

        return result;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize syntax highlighter
const highlighter = new SyntaxHighlighter();
highlighter.init();