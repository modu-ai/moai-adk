import 'nextra-theme-docs/style.css'
import '../styles/globals.css'
import { useEffect } from 'react'

export default function App({ Component, pageProps }) {
  useEffect(() => {
    // Nextra의 인라인 스타일을 오버라이드하여 무채색 테마 적용
    const overrideNextraColors = () => {
      // CSS 변수 직접 설정
      const root = document.documentElement
      root.style.setProperty('--nextra-primary-hue', '0deg')
      root.style.setProperty('--nextra-primary-saturation', '0%')
      
      // 다크 모드 감지
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
    
    // 짧은 간격으로 반복 실행 (Nextra가 동적으로 스타일을 주입할 수 있음)
    const interval = setInterval(overrideNextraColors, 100)
    
    // 다크 모드 변경 감지
    const observer = new MutationObserver(() => {
      setTimeout(overrideNextraColors, 50)
    })
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class', 'data-theme']
    })
    
    // DOM 변경 감지 (새로운 style 태그 추가 시)
    const domObserver = new MutationObserver(overrideNextraColors)
    domObserver.observe(document.head, {
      childList: true,
      subtree: true
    })
    
    return () => {
      clearInterval(interval)
      observer.disconnect()
      domObserver.disconnect()
    }
  }, [])
  
  return <Component {...pageProps} />
}


