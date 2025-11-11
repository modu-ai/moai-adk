#!/usr/bin/env node

console.log('🧪 Testing Nextra styling fixes...\n')

// Test 1: Check if critical files exist
const fs = require('fs')
const path = require('path')

const requiredFiles = [
  'pages/_app.tsx',
  'pages/_document.tsx',
  'theme.config.tsx',
  'styles/globals.css',
  'next.config.js'
]

console.log('📁 Checking required files:')
let allFilesExist = true

requiredFiles.forEach(file => {
  const exists = fs.existsSync(file)
  console.log(`  ${exists ? '✅' : '❌'} ${file}`)
  if (!exists) allFilesExist = false
})

if (!allFilesExist) {
  console.log('\n❌ Some required files are missing!')
  process.exit(1)
}

// Test 2: Check theme.config.tsx structure
console.log('\n🎨 Checking theme.config.tsx structure:')
try {
  const themeConfig = fs.readFileSync('theme.config.tsx', 'utf8')

  const checks = [
    { name: 'Exports config object', pattern: /export default config/ },
    { name: 'No missing imports', pattern: /CustomSearch|CoreWebVitalsOptimizer/ },
    { name: 'Has logo configuration', pattern: /logo:/ },
    { name: 'Has project configuration', pattern: /project:/ },
    { name: 'Has search configuration', pattern: /search:/ },
    { name: 'Has i18n configuration', pattern: /i18n:/ }
  ]

  checks.forEach(({ name, pattern }) => {
    const matches = pattern.test(themeConfig)
    if (name.includes('No missing')) {
      console.log(`  ${!matches ? '✅' : '❌'} ${name}`)
    } else {
      console.log(`  ${matches ? '✅' : '❌'} ${name}`)
    }
  })
} catch (error) {
  console.log('  ❌ Error reading theme.config.tsx:', error.message)
}

// Test 3: Check _app.tsx imports
console.log('\n📱 Checking _app.tsx imports:')
try {
  const appConfig = fs.readFileSync('pages/_app.tsx', 'utf8')

  const appChecks = [
    { name: 'Imports global CSS', pattern: /import.*globals\.css/ },
    { name: 'Imports Nextra theme CSS', pattern: /import.*nextra-theme-docs\/style\.css/ },
    { name: 'Has proper App component', pattern: /export default function App/ }
  ]

  appChecks.forEach(({ name, pattern }) => {
    const matches = pattern.test(appConfig)
    console.log(`  ${matches ? '✅' : '❌'} ${name}`)
  })
} catch (error) {
  console.log('  ❌ Error reading pages/_app.tsx:', error.message)
}

// Test 4: Check package.json dependencies
console.log('\n📦 Checking package.json dependencies:')
try {
  const packageJson = JSON.parse(fs.readFileSync('package.json', 'utf8'))

  const deps = [
    'next',
    'nextra',
    'nextra-theme-docs',
    'react',
    'react-dom'
  ]

  deps.forEach(dep => {
    const exists = packageJson.dependencies?.[dep]
    console.log(`  ${exists ? '✅' : '❌'} ${dep}@${exists || 'missing'}`)
  })
} catch (error) {
  console.log('  ❌ Error reading package.json:', error.message)
}

console.log('\n🎯 Manual testing steps:')
console.log('1. Run: npm run dev')
console.log('2. Open: http://localhost:3000')
console.log('3. Check for:')
console.log('   - Modern fonts (not Times New Roman)')
console.log('   - Navigation sidebar')
console.log('   - Search functionality')
console.log('   - Dark/light mode toggle')
console.log('   - Proper styling and layout')

console.log('\n✨ If styling still doesn\'t work, try:')
console.log('   - Delete .next directory and restart')
console.log('   - Clear browser cache')
console.log('   - Check browser console for errors')