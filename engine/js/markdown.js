(function(includeImgs) {
    // ==================== Node Type Definitions ====================
    // Defines all AST node types used in the markdown conversion
    const NodeType = {
        DOCUMENT: 'document',      // Root container
        HEADING: 'heading',        // h1-h6 → level: 1-6
        PARAGRAPH: 'paragraph',    // p, generic block container
        TEXT: 'text',              // Plain text content
        LINK: 'link',              // a → href, children
        STRONG: 'strong',          // strong, b
        EMPHASIS: 'emphasis',      // em, i
        UNDERLINE: 'underline',    // u
        STRIKETHROUGH: 'strike',   // s, strike, del
        CODE: 'code',              // Inline code
        CODE_BLOCK: 'codeblock',   // pre/code block → lang, content
        LIST: 'list',              // ul/ol → ordered: bool
        LIST_ITEM: 'listitem',     // li
        BLOCKQUOTE: 'blockquote',  // blockquote
        IMAGE: 'image',            // img → src, alt
        BREAK: 'break',            // br
        RULE: 'rule',              // hr
    };

    // ==================== AST Node Factory ====================
    /**
     * Creates a new AST node with the given type and attributes
     * @param {string} type - The node type from NodeType
     * @param {Object} attrs - Type-specific attributes (href, level, lang, etc.)
     * @returns {Object} The new AST node
     */
    function createNode(type, attrs = {}) {
        return {
            type: type,
            children: [],
            content: '',
            attrs: attrs
        };
    }

    // ==================== DOM Visibility Checks ====================
    // Check element visibility using ComputedStyle
    function isVisible(el) {
        if (!el || el.nodeType !== Node.ELEMENT_NODE) return true;
        const style = window.getComputedStyle(el);
        return style.display !== 'none' &&
               style.visibility !== 'hidden' &&
               parseFloat(style.opacity) !== 0;
    }

    // Find the best content root (main > article > body)
    function getContentRoot() {
        const main = document.querySelector('main');
        if (main && isVisible(main)) return main;

        const article = document.querySelector('article');
        if (article && isVisible(article)) return article;

        return document.body;
    }

    // ==================== DOM → AST Conversion ====================
    /**
     * Builds an AST from the DOM tree
     * @param {Element} element - The root element to convert
     * @returns {Object} The root AST node
     */
    function buildAST(element) {
        const root = createNode(NodeType.DOCUMENT);
        const children = traverseDOM(element);
        root.children = children;
        return root;
    }

    /**
     * Recursively traverses the DOM and converts elements to AST nodes
     * @param {Node} node - The current DOM node
     * @returns {Array} Array of AST nodes
     */
    function traverseDOM(node) {
        const nodes = [];

        if (node.nodeType === Node.TEXT_NODE) {
            const text = node.textContent.trim();
            if (text.length > 0) {
                const textNode = createNode(NodeType.TEXT);
                textNode.content = text;
                nodes.push(textNode);
            }
            return nodes;
        }

        if (node.nodeType !== Node.ELEMENT_NODE) {
            return nodes;
        }

        if (!isVisible(node)) {
            return nodes;
        }

        const tag = node.tagName?.toLowerCase();

        // Skip non-content elements
        const skipTags = ['script', 'style', 'noscript', 'nav', 'header', 'footer', 'aside', 'form', 'input', 'button', 'select', 'textarea', 'iframe', 'svg', 'canvas'];
        if (skipTags.includes(tag)) {
            return nodes;
        }

        // Handle specific semantic tags
        let astNode = null;

        switch (tag) {
            case 'h1':
            case 'h2':
            case 'h3':
            case 'h4':
            case 'h5':
            case 'h6':
                astNode = createNode(NodeType.HEADING, { level: parseInt(tag[1]) });
                astNode.children = traverseChildren(node);
                nodes.push(astNode);
                break;

            case 'p':
                astNode = createNode(NodeType.PARAGRAPH);
                astNode.children = traverseChildren(node);
                // Only add non-empty paragraphs
                if (astNode.children.length > 0) {
                    nodes.push(astNode);
                }
                break;

            case 'a':
                astNode = createNode(NodeType.LINK, { href: node.getAttribute('href') || '' });
                astNode.children = traverseChildren(node);
                // Only add non-empty links
                if (astNode.children.length > 0) {
                    nodes.push(astNode);
                }
                break;

            case 'strong':
            case 'b':
                astNode = createNode(NodeType.STRONG);
                astNode.children = traverseChildren(node);
                if (astNode.children.length > 0) {
                    nodes.push(astNode);
                }
                break;

            case 'em':
            case 'i':
                astNode = createNode(NodeType.EMPHASIS);
                astNode.children = traverseChildren(node);
                if (astNode.children.length > 0) {
                    nodes.push(astNode);
                }
                break;

            case 'u':
                astNode = createNode(NodeType.UNDERLINE);
                astNode.children = traverseChildren(node);
                if (astNode.children.length > 0) {
                    nodes.push(astNode);
                }
                break;

            case 's':
            case 'strike':
            case 'del':
                astNode = createNode(NodeType.STRIKETHROUGH);
                astNode.children = traverseChildren(node);
                if (astNode.children.length > 0) {
                    nodes.push(astNode);
                }
                break;

            case 'code':
                // Only treat as inline code if not inside a pre block
                if (node.closest('pre') === null) {
                    astNode = createNode(NodeType.CODE);
                    astNode.content = node.textContent;
                    nodes.push(astNode);
                } else {
                    // Inside pre, just traverse children
                    nodes.push(...traverseChildren(node));
                }
                break;

            case 'pre':
                const lang = node.dataset.language || '';
                let preTxt = node.innerText?.trim() || '';
                if (preTxt === '') {
                    preTxt = node.textContent.trim();
                }
                astNode = createNode(NodeType.CODE_BLOCK, { lang: lang });
                astNode.content = preTxt;
                nodes.push(astNode);
                break;

            case 'ul':
                astNode = createNode(NodeType.LIST, { ordered: false });
                astNode.children = traverseListItems(node);
                if (astNode.children.length > 0) {
                    nodes.push(astNode);
                }
                break;

            case 'ol':
                astNode = createNode(NodeType.LIST, { ordered: true });
                astNode.children = traverseListItems(node);
                if (astNode.children.length > 0) {
                    nodes.push(astNode);
                }
                break;

            case 'li':
                // li nodes are handled by traverseListItems
                nodes.push(...traverseChildren(node));
                break;

            case 'blockquote':
                astNode = createNode(NodeType.BLOCKQUOTE);
                astNode.children = traverseChildren(node);
                if (astNode.children.length > 0) {
                    nodes.push(astNode);
                }
                break;

            case 'img':
                if (includeImgs) {
                    astNode = createNode(NodeType.IMAGE, {
                        src: node.getAttribute('src') || '',
                        alt: node.getAttribute('alt') || ''
                    });
                    nodes.push(astNode);
                }
                break;

            case 'br':
                astNode = createNode(NodeType.BREAK);
                nodes.push(astNode);
                break;

            case 'hr':
                astNode = createNode(NodeType.RULE);
                nodes.push(astNode);
                break;

            default:
                // For other elements, just traverse children
                nodes.push(...traverseChildren(node));
                break;
        }

        return nodes;
    }

    /**
     * Traverses all child nodes of an element
     * @param {Element} element - The parent element
     * @returns {Array} Array of AST nodes from all children
     */
    function traverseChildren(element) {
        const nodes = [];
        for (const child of element.childNodes) {
            nodes.push(...traverseDOM(child));
        }
        return nodes;
    }

    /**
     * Special handler for list items - ensures proper nesting
     * @param {Element} listElement - The ul or ol element
     * @returns {Array} Array of list item AST nodes
     */
    function traverseListItems(listElement) {
        const items = [];
        for (const child of listElement.children) {
            if (child.tagName?.toLowerCase() === 'li' && isVisible(child)) {
                const itemNode = createNode(NodeType.LIST_ITEM);
                itemNode.children = traverseChildren(child);
                // Preserve nested lists by checking for ul/ol in children
                if (itemNode.children.length > 0) {
                    items.push(itemNode);
                }
            }
        }
        return items;
    }

    // ==================== AST → Markdown Rendering ====================
    /**
     * Renders an AST to markdown string with proper whitespace handling
     * @param {Object} ast - The root AST node
     * @returns {string} The rendered markdown
     */
    function renderAST(ast) {
        const lines = [];
        renderNode(ast, lines);
        const result = lines.join('');

        // Final cleanup: collapse excessive newlines to max 2
        return result.replace(/\n{3,}/g, '\n\n').trim();
    }

    /**
     * Ensures a block-level element starts on a new line
     * Block elements should be preceded by at least one newline character.
     * If the last character in the output buffer is not a newline, add newlines.
     * @param {Array} lines - Accumulator for output lines
     */
    function ensureBlockStart(lines) {
        if (lines.length === 0) return;
        
        // Join all lines to get the current output
        const output = lines.join('');
        
        // If output is empty or already ends with newline, no need to add more
        if (output.length === 0 || output.endsWith('\n\n')) {
            return;
        }
        
        // If output ends with one newline, add one more for proper spacing
        if (output.endsWith('\n')) {
            lines.push('\n');
        } else {
            // If output doesn't end with newline, add two newlines
            lines.push('\n\n');
        }
    }

    /**
     * Renders a single AST node
     * @param {Object} node - The AST node to render
     * @param {Array} lines - Accumulator for output lines
     */
    function renderNode(node, lines) {
        if (!node) return;

        switch (node.type) {
            case NodeType.DOCUMENT:
                renderChildren(node.children, lines);
                break;

            case NodeType.HEADING:
                ensureBlockStart(lines);
                const hashes = '#'.repeat(node.attrs.level);
                const headingText = renderInlineChildren(node.children);
                lines.push(hashes + ' ' + headingText + '\n\n');
                break;

            case NodeType.PARAGRAPH:
                ensureBlockStart(lines);
                const paraText = renderInlineChildren(node.children);
                if (paraText.trim().length > 0) {
                    lines.push(paraText + '\n\n');
                }
                break;

            case NodeType.TEXT:
                const cleanedText = node.content.replace(/\s+/g, ' ').trim();
                if (cleanedText.length > 0) {
                    lines.push(cleanedText);
                }
                break;

            case NodeType.LINK:
                const linkText = renderInlineChildren(node.children);
                if (linkText.trim().length > 0) {
                    lines.push('[' + linkText + '](' + node.attrs.href + ')');
                }
                break;

            case NodeType.STRONG:
                const strongText = renderInlineChildren(node.children);
                if (strongText.trim().length > 0) {
                    lines.push('**' + strongText + '**');
                }
                break;

            case NodeType.EMPHASIS:
                const emphasisText = renderInlineChildren(node.children);
                if (emphasisText.trim().length > 0) {
                    lines.push('*' + emphasisText + '*');
                }
                break;

            case NodeType.UNDERLINE:
                const underlineText = renderInlineChildren(node.children);
                if (underlineText.trim().length > 0) {
                    lines.push('<u>' + underlineText + '</u>');
                }
                break;

            case NodeType.STRIKETHROUGH:
                const strikeText = renderInlineChildren(node.children);
                if (strikeText.trim().length > 0) {
                    lines.push('~~' + strikeText + '~~');
                }
                break;

            case NodeType.CODE:
                const codeText = node.content.replace(/\s+/g, ' ').trim();
                if (codeText.length > 0) {
                    lines.push('`' + codeText + '`');
                }
                break;

            case NodeType.CODE_BLOCK:
                ensureBlockStart(lines);
                const lang = node.attrs.lang || '';
                lines.push('```' + lang + '\n' + node.content + '\n```\n\n');
                break;

            case NodeType.LIST:
                ensureBlockStart(lines);
                renderList(node, lines);
                break;

            case NodeType.LIST_ITEM:
                // List items are rendered by renderList
                break;

            case NodeType.BLOCKQUOTE:
                ensureBlockStart(lines);
                const quoteLines = [];
                renderChildren(node.children, quoteLines);
                const quoteText = quoteLines.join('').trim();
                const blockquoteLines = quoteText.split('\n');
                for (const line of blockquoteLines) {
                    lines.push('> ' + line + '\n');
                }
                lines.push('\n');
                break;

            case NodeType.IMAGE:
                lines.push('![' + node.attrs.alt + '](' + node.attrs.src + ')\n');
                break;

            case NodeType.BREAK:
                lines.push('\n');
                break;

            case NodeType.RULE:
                ensureBlockStart(lines);
                lines.push('---\n\n');
                break;
        }
    }

    /**
     * Renders a list (ul/ol) with proper item numbering
     * @param {Object} listNode - The LIST node
     * @param {Array} lines - Accumulator for output lines
     */
    function renderList(listNode, lines) {
        let itemIndex = 1;
        for (const itemNode of listNode.children) {
            if (itemNode.type === NodeType.LIST_ITEM) {
                // For list items, render inline content first
                let itemText = '';
                const itemLines = [];
                
                // First pass: check if there are block-level elements
                let hasBlockElements = false;
                for (const child of itemNode.children) {
                    if (child.type === NodeType.LIST || child.type === NodeType.BLOCKQUOTE || 
                        child.type === NodeType.CODE_BLOCK) {
                        hasBlockElements = true;
                        break;
                    }
                }
                
                if (hasBlockElements) {
                    // Render block elements, collecting output
                    renderChildren(itemNode.children, itemLines);
                    itemText = itemLines.join('').trim();
                } else {
                    // Simple case: just inline content
                    itemText = renderInlineChildren(itemNode.children);
                }
                
                const prefix = listNode.attrs.ordered ? (itemIndex++ + '. ') : '- ';
                const itemContentLines = itemText.split('\n');
                
                // Add first line with prefix
                if (itemContentLines.length > 0) {
                    lines.push(prefix + itemContentLines[0] + '\n');
                    
                    // Add remaining lines with indentation
                    for (let i = 1; i < itemContentLines.length; i++) {
                        if (itemContentLines[i].trim().length > 0) {
                            lines.push('  ' + itemContentLines[i] + '\n');
                        } else if (i < itemContentLines.length - 1) {
                            lines.push('\n');
                        }
                    }
                }
            }
        }
        lines.push('\n');
    }

    /**
     * Renders inline children (for use within block elements)
     * Builds a string without adding block-level spacing
     * @param {Array} children - Array of child nodes
     * @returns {string} The rendered inline content
     */
    function renderInlineChildren(children) {
        const result = [];
        for (const child of children) {
            if (child.type === NodeType.TEXT) {
                const cleanedText = child.content.replace(/\s+/g, ' ').trim();
                if (cleanedText.length > 0) {
                    if (result.length > 0 && !result[result.length - 1].endsWith(' ')) {
                        result.push(' ');
                    }
                    result.push(cleanedText);
                }
            } else if (child.type === NodeType.LINK) {
                const linkText = renderInlineChildren(child.children);
                if (linkText.trim().length > 0) {
                    if (result.length > 0 && !result[result.length - 1].endsWith(' ')) {
                        result.push(' ');
                    }
                    result.push('[' + linkText + '](' + child.attrs.href + ')');
                }
            } else if (child.type === NodeType.STRONG) {
                const strongText = renderInlineChildren(child.children);
                if (strongText.trim().length > 0) {
                    result.push('**' + strongText + '**');
                }
            } else if (child.type === NodeType.EMPHASIS) {
                const emphasisText = renderInlineChildren(child.children);
                if (emphasisText.trim().length > 0) {
                    result.push('*' + emphasisText + '*');
                }
            } else if (child.type === NodeType.UNDERLINE) {
                const underlineText = renderInlineChildren(child.children);
                if (underlineText.trim().length > 0) {
                    result.push('<u>' + underlineText + '</u>');
                }
            } else if (child.type === NodeType.STRIKETHROUGH) {
                const strikeText = renderInlineChildren(child.children);
                if (strikeText.trim().length > 0) {
                    result.push('~~' + strikeText + '~~');
                }
            } else if (child.type === NodeType.CODE) {
                const codeText = child.content.replace(/\s+/g, ' ').trim();
                if (codeText.length > 0) {
                    result.push('`' + codeText + '`');
                }
            } else if (child.type === NodeType.IMAGE) {
                result.push('![' + child.attrs.alt + '](' + child.attrs.src + ')');
            } else if (child.type === NodeType.BREAK) {
                result.push('\n');
            }
        }
        return result.join('');
    }

    /**
     * Renders children nodes (block-level, for use at document level)
     * @param {Array} children - Array of child nodes
     * @param {Array} lines - Accumulator for output lines
     */
    function renderChildren(children, lines) {
        for (const child of children) {
            renderNode(child, lines);
        }
    }

    // ==================== Main Execution ====================
    const root = getContentRoot();
    const ast = buildAST(root);
    const result = renderAST(ast);

    return result;
})(%t)
