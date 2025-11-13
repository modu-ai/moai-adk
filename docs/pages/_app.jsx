import 'nextra-theme-docs/style.css'
import '../styles/globals.css'
import { useEffect } from 'react'

export default function App({ Component, pageProps }) {
  useEffect(() => {
    // Nextra의 인라인 스타일을 오버라이드하여 무채색 테마 적용
    // 클릭 문제를 해결하기 위해 setInterval 제거하고 MutationObserver만 사용
    const overrideNextraColors = () => {
      const root = document.documentElement
      root.style.setProperty('--nextra-primary-hue', '0deg')
      root.style.setProperty('--nextra-primary-saturation', '0%')
      
      const isDark = document.documentElement.classList.contains('dark') || 
                     document.documentElement.getAttribute('data-theme') === 'dark'
      
      if (isDark) {
        root.style.setProperty('--nextra-primary-lightness', '100%')
      } else {
        root.style.setProperty('--nextra-primary-lightness', '0%')
      }
      
      // <style> 태그에서도 오버라이드
      const styleTags = document.querySelectorAll('style')
      styleTags.forEach(style => {
        if (style.textContent && style.textContent.includes('--nextra-primary-hue')) {
          const newContent = style.textContent
            .replace(/--nextra-primary-hue:\s*\d+deg/g, '--nextra-primary-hue: 0deg')
            .replace(/--nextra-primary-saturation:\s*\d+%/g, '--nextra-primary-saturation: 0%')
            .replace(/--nextra-primary-lightness:\s*\d+%/g, isDark ? '--nextra-primary-lightness: 100%' : '--nextra-primary-lightness: 0%')
          style.textContent = newContent
        }
      })
    }
    
    // 초기 실행
    overrideNextraColors()
    
    // 다크 모드 변경 감지 (debounce 적용)
    let timeoutId = null
    const observer = new MutationObserver(() => {
      if (timeoutId) clearTimeout(timeoutId)
      timeoutId = setTimeout(overrideNextraColors, 100)
    })
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class', 'data-theme']
    })
    
    // DOM 변경 감지 (새로운 style 태그 추가 시, debounce 적용)
    const domObserver = new MutationObserver(() => {
      if (timeoutId) clearTimeout(timeoutId)
      timeoutId = setTimeout(overrideNextraColors, 100)
    })
    domObserver.observe(document.head, {
      childList: true,
      subtree: true
    })
    
    return () => {
      if (timeoutId) clearTimeout(timeoutId)
      observer.disconnect()
      domObserver.disconnect()
    }
  }, [])
  
  return <Component {...pageProps} />
}


