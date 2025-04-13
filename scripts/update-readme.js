// Script to update README.md with screenshots
const fs = require('fs');
const path = require('path');

// Configuration
const README_PATH = path.join(__dirname, '../README.md');
const SCREENSHOTS_DIR = path.join(__dirname, '../docs/screenshots');
const SCREENSHOTS_SECTION_START = '<!-- SCREENSHOTS_START -->';
const SCREENSHOTS_SECTION_END = '<!-- SCREENSHOTS_END -->';
const DEBUG_MODE = false; // Set to false to use real screenshots

// Screenshots configuration
const SCREENSHOTS = [
  {
    title: 'Dashboard',
    description: 'Overview of residents, payments, and expenses with visual indicators.',
    filename: 'dashboard.png'
  },
  {
    title: 'Residents',
    description: 'Manage condo residents with search and filtering capabilities.',
    filename: 'residents.png'
  },
  {
    title: 'Payments',
    description: 'Track payments with detailed information and filtering.',
    filename: 'payments.png'
  },
  {
    title: 'Expenses',
    description: 'Record and categorize expenses with search functionality.',
    filename: 'expenses.png'
  },
  {
    title: 'Reports',
    description: 'Visual reports showing payment trends and expense breakdowns.',
    filename: 'reports.png'
  }
];

// Ensure screenshots directory exists
if (!fs.existsSync(SCREENSHOTS_DIR)) {
  fs.mkdirSync(SCREENSHOTS_DIR, { recursive: true });
}

// Create placeholder images in debug mode
if (DEBUG_MODE) {
  console.log('Debug mode: Creating placeholder screenshots...');
  SCREENSHOTS.forEach(screenshot => {
    const filePath = path.join(SCREENSHOTS_DIR, screenshot.filename);
    if (!fs.existsSync(filePath)) {
      const content = `DEBUG MODE: Screenshot of ${screenshot.title} page - ${new Date().toISOString()}`;
      fs.writeFileSync(filePath, content);
      console.log(`Created placeholder for ${screenshot.title}`);
    }
  });
}

function updateReadme() {
  console.log('Updating README.md with screenshots...');
  
  // Read README content
  let readmeContent = fs.readFileSync(README_PATH, 'utf8');
  
  // Check if screenshots section markers exist
  if (!readmeContent.includes(SCREENSHOTS_SECTION_START) || !readmeContent.includes(SCREENSHOTS_SECTION_END)) {
    console.log('Adding screenshots section markers to README.md');
    // Append screenshots section at the end
    readmeContent += `\n\n## Screenshots\n\n${SCREENSHOTS_SECTION_START}\n${SCREENSHOTS_SECTION_END}\n`;
  }
  
  // Generate screenshots markdown
  let screenshotsContent = `\n## Screenshots\n\n`;
  
  for (const screenshot of SCREENSHOTS) {
    const imagePath = `docs/screenshots/${screenshot.filename}`;
    
    // Check if screenshot exists
    const fullImagePath = path.join(__dirname, '..', imagePath);
    const fileExists = fs.existsSync(fullImagePath);
    
    screenshotsContent += `### ${screenshot.title}\n\n`;
    screenshotsContent += `${screenshot.description}\n\n`;
    
    if (fileExists) {
      screenshotsContent += `![${screenshot.title}](${imagePath})\n\n`;
    } else if (DEBUG_MODE) {
      screenshotsContent += `_[Debug mode: Screenshot placeholder for ${screenshot.title}]_\n\n`;
    } else {
      screenshotsContent += `_[Screenshot not available]_\n\n`;
    }
  }
  
  // Update screenshots section in README
  const updatedContent = readmeContent.replace(
    new RegExp(`${SCREENSHOTS_SECTION_START}[\\s\\S]*?${SCREENSHOTS_SECTION_END}`),
    `${SCREENSHOTS_SECTION_START}\n${screenshotsContent}\n${SCREENSHOTS_SECTION_END}`
  );
  
  // Write updated content to README
  fs.writeFileSync(README_PATH, updatedContent);
  
  console.log('README.md updated with screenshots');
}

updateReadme(); 