(function() {
  const links = [];
  const anchorElements = document.querySelectorAll('a');
  
  for (const anchor of anchorElements) {
    const href = anchor.getAttribute('href');
    const text = anchor.textContent.trim();
    
    // Skip anchors without href or text
    if (href && text) {
      links.push({ text, href });
    }
  }
  
  return JSON.stringify(links);
})()
