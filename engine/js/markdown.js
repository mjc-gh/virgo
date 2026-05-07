(function(includeImgs) {
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

    // Get href attribute for anchor elements
    function getHref(el) {
        return el.getAttribute('href') || '';
    }

    // Get text content trimmed
    // Process child nodes recursively
    function processChildren(el) {
        let result = '';
        for (const child of el.childNodes) {
            if (child.nodeType === Node.TEXT_NODE) {
                result += child.textContent;
            } else if (child.nodeType === Node.ELEMENT_NODE) {
                result += convertElement(child);
            }
        }
        return result;
    }

    // Convert list elements (ul/ol)
    function convertList(el, ordered) {
        let result = '';
        let index = 1;
        for (const child of el.children) {
            if (child.tagName?.toLowerCase() === 'li' && isVisible(child)) {
                const prefix = ordered ? (index++ + '. ') : '- ';
                result += prefix + processChildren(child).trim() + '\n';
            }
        }
        return result;
    }

    // Convert blockquote with > prefix on each line
    function convertBlockquote(el) {
        const content = processChildren(el).trim();
        return content.split('\n').map(line => '> ' + line).join('\n');
    }

    function squishText(txt) {
        return txt.replace(/\s+/g, ' ').trim();
    }

    // Recursive element conversion
    function convertElement(el) {
        if (!isVisible(el)) return '';

        const tag = el.tagName?.toLowerCase();

        // Handle specific tags
        switch (tag) {
            // Headers
            case 'h1': return '# ' + squishText(processChildren(el)) + '\n\n';
            case 'h2': return '## ' + squishText(processChildren(el)) + '\n\n';
            case 'h3': return '### ' + squishText(processChildren(el)) + '\n\n';
            case 'h4': return '#### ' + squishText(processChildren(el)) + '\n\n';
            case 'h5': return '##### ' + squishText(processChildren(el)) + '\n\n';
            case 'h6': return '###### ' + squishText(processChildren(el)) + '\n\n';

            // Links
            case 'a':
                const anchorTxt = squishText(processChildren(el))
                if (anchorTxt === "") {
                    return '';
                }

                return '[' + anchorTxt + '](' + getHref(el) + ') ';

            // Text formatting
            case 'strong':
            case 'b':
                return '**' + processChildren(el) + '** ';

            case 'em':
            case 'i':
                return '*' + processChildren(el) + '* ';

            case 'u':
                return '<u>' + processChildren(el) + '</u> ';

            case 's':
            case 'strike':
            case 'del':
                return '~~' + processChildren(el) + '~~ ';

            // Paragraphs and line breaks
            case 'p':
                pTxt = squishText(processChildren(el));
                if (pTxt === "")
                    return "";

                return pTxt + '\n\n';

            case 'br':
                return '\n';

            case 'hr':
                return '---\n\n';

            // Lists
            case 'ul':
                return convertList(el, false) + '\n\n';

            case 'ol':
                return convertList(el, true) + '\n\n';

            case 'li':
                return processChildren(el).trim() + '\n';

            // Blockquotes
            case 'blockquote':
                return convertBlockquote(el) + '\n\n';

            // Code
            case 'code':
                if (el.parentElement?.tagName?.toLowerCase() === 'pre') {
                    return el.textContent;
                }
                return '`' + el.textContent + '` ';

            case 'pre':
                return '```\n' + el.textContent.trim() + '\n```\n\n';

            // Images
            case 'img':
                if (!includeImgs) return '';
                const alt = el.getAttribute('alt') || '';
                const src = el.getAttribute('src') || '';
                return '![' + alt + '](' + src + ')\n';

            // Skip non-content elements
            case 'script':
            case 'style':
            case 'noscript':
            case 'nav':
            case 'header':
            case 'footer':
            case 'aside':
            case 'form':
            case 'input':
            case 'button':
            case 'select':
            case 'textarea':
            case 'iframe':
            case 'svg':
            case 'canvas':
                return '';

            default:
                return processChildren(el).trim() + ' ';
        }
    }

    // Main execution
    const root = getContentRoot();
    const result = processChildren(root).trim();

    // Final clean up
    // - Replace excessive newlines (more than 2 consecutive)
    // - Replace excessive blankspace before headers
    return result
        .replace(/\n[ ]+(#{1,6})/g, '\n$1')
        .replace(/\n{3,}/g, '\n\n');
})(%t)
