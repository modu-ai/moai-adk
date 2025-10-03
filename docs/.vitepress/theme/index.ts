// Custom theme extending default VitePress theme
import DefaultTheme from 'vitepress/theme'
import './custom.css'
import type { Theme } from 'vitepress'

// Lucide Icons - 무채색 아이콘 시스템
import {
  Package,
  Ruler,
  Tag,
  CheckCircle,
  RefreshCw,
  Globe,
  Sparkles,
  Workflow,
  Wrench,
  Users,
  Zap
} from 'lucide-vue-next'

export default {
  ...DefaultTheme,
  enhanceApp({ app }) {
    // Lucide Icons 글로벌 등록 (무채색 zinc 테마)
    app.component('IconPackage', Package)
    app.component('IconRuler', Ruler)
    app.component('IconTag', Tag)
    app.component('IconCheck', CheckCircle)
    app.component('IconRefresh', RefreshCw)
    app.component('IconGlobe', Globe)
    app.component('IconSparkles', Sparkles)
    app.component('IconWorkflow', Workflow)
    app.component('IconWrench', Wrench)
    app.component('IconUsers', Users)
    app.component('IconZap', Zap)
  }
} as Theme