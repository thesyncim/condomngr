// Screenshot capture script using Puppeteer
const puppeteer = require('puppeteer');
const fs = require('fs');
const path = require('path');

// Configuration
const SCREENSHOTS_DIR = path.join(__dirname, '../docs/screenshots');
const BASE_URL = 'http://localhost:8080';
const WAIT_TIME = 1000; // Time to wait for page to load properly in ms

// Pages to capture
const PAGES = [
  { name: 'dashboard', path: '#dashboard', filename: 'dashboard.png', fullPage: true },
  { name: 'residents', path: '#residents', filename: 'residents.png', fullPage: true },
  { name: 'payments', path: '#payments', filename: 'payments.png', fullPage: true },
  { name: 'expenses', path: '#expenses', filename: 'expenses.png', fullPage: true },
  { name: 'reports', path: '#reports', filename: 'reports.png', fullPage: true }
];

// Ensure screenshots directory exists
if (!fs.existsSync(SCREENSHOTS_DIR)) {
  fs.mkdirSync(SCREENSHOTS_DIR, { recursive: true });
}

async function captureScreenshots() {
  console.log('Launching browser...');
  let browser = null;
  
  try {
    // Check if server is accessible
    try {
      const response = await fetch(`${BASE_URL}`);
      if (!response.ok) {
        throw new Error(`Server responded with status: ${response.status}`);
      }
    } catch (error) {
      console.error(`Error connecting to server at ${BASE_URL}: ${error.message}`);
      process.exit(1);
    }
    
    browser = await puppeteer.launch({ 
      headless: 'new',
      defaultViewport: { width: 1280, height: 800 }
    });
    
    const page = await browser.newPage();
    
    for (const pageConfig of PAGES) {
      console.log(`Capturing ${pageConfig.name}...`);
      
      try {
        // Navigate to the page
        await page.goto(`${BASE_URL}/${pageConfig.path}`, {
          waitUntil: 'networkidle0',
          timeout: 30000
        });
        
        // Extra wait to ensure JS-rendered content appears
        await page.waitForTimeout(WAIT_TIME);

        // Take screenshot
        const screenshotPath = path.join(SCREENSHOTS_DIR, pageConfig.filename);
        await page.screenshot({
          path: screenshotPath,
          fullPage: pageConfig.fullPage
        });
        
        console.log(`Screenshot saved to ${screenshotPath}`);
      } catch (pageError) {
        console.error(`Error capturing ${pageConfig.name}: ${pageError.message}`);
        // Continue with other screenshots
      }
    }
  } catch (error) {
    console.error('Error capturing screenshots:', error);
    process.exit(1);
  } finally {
    if (browser) {
      await browser.close();
    }
    console.log('Screenshot capture completed');
  }
}

// Add proper error handling for the main function
captureScreenshots().catch(error => {
  console.error('Fatal error:', error);
  process.exit(1);
}); 