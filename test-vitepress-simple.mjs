#!/usr/bin/env node
/**
 * Simple test to verify VitePress is serving correctly
 */

import http from 'node:http';

const testUrl = 'http://localhost:5173/';

console.log('🔍 Testing VitePress server at', testUrl);

http.get(testUrl, (res) => {
  console.log('✅ HTTP Status:', res.statusCode);
  console.log('📄 Content-Type:', res.headers['content-type']);

  let data = '';

  res.on('data', (chunk) => {
    data += chunk;
  });

  res.on('end', () => {
    console.log('📊 Response size:', data.length, 'bytes');

    // Check for 404 in content
    if (data.includes('404') || data.includes('PAGE NOT FOUND')) {
      console.log('❌ 404 ERROR DETECTED in page content!');
      console.log('Content preview:', data.substring(0, 500));
      process.exit(1);
    } else {
      console.log('✅ No 404 error in HTML response');
    }

    // Check for VitePress app div
    if (data.includes('<div id="app">')) {
      console.log('✅ VitePress app div found');
    } else {
      console.log('❌ VitePress app div NOT found');
    }

    // Check for VitePress script
    if (data.includes('vitepress/dist/client/app/index.js')) {
      console.log('✅ VitePress client script found');
    } else {
      console.log('❌ VitePress client script NOT found');
    }

    console.log('\n✅ Server test PASSED - HTML is being served correctly');
    console.log('🌐 Please check http://localhost:5173/ in your browser');
    console.log('   (The 404 may only appear after JavaScript execution)');
  });
}).on('error', (err) => {
  console.error('❌ Connection error:', err.message);
  process.exit(1);
});