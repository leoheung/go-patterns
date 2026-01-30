import { defineConfig } from 'vitepress'

export default defineConfig({
  base: '/go-patterns/',
  lastUpdated: true,
  cleanUrls: true,
  metaChunk: true,

  locales: {
    'en': {
      label: 'English',
      lang: 'en-US',
      link: '/en/',
      title: 'Go Patterns',
      description: 'Go concurrency patterns and data structures',
      themeConfig: {
        nav: nav('en'),
        sidebar: sidebar('en'),
        editLink: {
          pattern: 'https://github.com/leoxiang66/go-patterns/edit/main/doc/:path',
          text: 'Edit this page on GitHub'
        }
      }
    },
    'zh-hk': {
      label: '繁體中文',
      lang: 'zh-HK',
      link: '/zh-hk/',
      title: 'Go Patterns',
      description: 'Go 並發模式與數據結構',
      themeConfig: {
        nav: nav('zh-hk'),
        sidebar: sidebar('zh-hk'),
        editLink: {
          pattern: 'https://github.com/leoxiang66/go-patterns/edit/main/doc/:path',
          text: '在 GitHub 上編輯此頁面'
        }
      }
    }
  },

  themeConfig: {
    socialLinks: [
      { icon: 'github', link: 'https://github.com/leoxiang66/go-patterns' }
    ],

    search: {
      provider: 'local'
    },

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024-present'
    }
  }
})

function nav(lang: string) {
  const isEn = lang === 'en'
  return [
    {
      text: isEn ? 'Guide' : '指南',
      link: `/${lang}/guide/getting-started`,
      activeMatch: `/${lang}/guide/`
    },
    {
      text: isEn ? 'Container' : '容器',
      link: `/${lang}/container/`,
      activeMatch: `/${lang}/container/`
    },
    {
      text: isEn ? 'Parallel' : '並發',
      link: `/${lang}/parallel/`,
      activeMatch: `/${lang}/parallel/`
    },
    {
      text: isEn ? 'Utils' : '工具',
      link: `/${lang}/utils/`,
      activeMatch: `/${lang}/utils/`
    },
    {
      text: isEn ? 'Cryptography' : '加密',
      link: `/${lang}/cryptography/`,
      activeMatch: `/${lang}/cryptography/`
    },
    {
      text: isEn ? 'Net' : '網絡',
      link: `/${lang}/net/`,
      activeMatch: `/${lang}/net/`
    }
  ]
}

function sidebar(lang: string) {
  const isEn = lang === 'en'
  return {
    [`/${lang}/guide/`]: [
      {
        text: isEn ? 'Getting Started' : '快速開始',
        items: [
          { text: isEn ? 'Introduction' : '介紹', link: `/${lang}/guide/getting-started` }
        ]
      }
    ],
    [`/${lang}/container/`]: [
      {
        text: isEn ? 'Container' : '容器',
        items: [
          { text: isEn ? 'Overview' : '概覽', link: `/${lang}/container/` },
          { text: 'List', link: `/${lang}/container/list` },
          { text: isEn ? 'Message Queue' : '消息隊列', link: `/${lang}/container/msgqueue` },
          { text: isEn ? 'Priority Queue' : '優先隊列', link: `/${lang}/container/pq` },
          { text: 'Cache', link: `/${lang}/container/cache` }
        ]
      }
    ],
    [`/${lang}/parallel/`]: [
      {
        text: isEn ? 'Parallel' : '並發',
        items: [
          { text: isEn ? 'Overview' : '概覽', link: `/${lang}/parallel/` },
          { text: 'Barrier', link: `/${lang}/parallel/barrier` },
          { text: isEn ? 'Limiter' : '限流器', link: `/${lang}/parallel/limiter` },
          { text: 'Mutex', link: `/${lang}/parallel/mutex` },
          { text: 'Pipeline', link: `/${lang}/parallel/pipeline` },
          { text: isEn ? 'Worker Pool' : '工作池', link: `/${lang}/parallel/pool` },
          { text: 'PubSub', link: `/${lang}/parallel/pubsub` },
          { text: isEn ? 'Read-Write Lock' : '讀寫鎖', link: `/${lang}/parallel/rwlock` },
          { text: 'Semaphore', link: `/${lang}/parallel/semaphore` }
        ]
      }
    ],
    [`/${lang}/utils/`]: [
      {
        text: isEn ? 'Utils' : '工具',
        items: [
          { text: isEn ? 'Overview' : '概覽', link: `/${lang}/utils/` }
        ]
      }
    ],
    [`/${lang}/cryptography/`]: [
      {
        text: isEn ? 'Cryptography' : '加密',
        items: [
          { text: isEn ? 'Overview' : '概覽', link: `/${lang}/cryptography/` }
        ]
      }
    ],
    [`/${lang}/net/`]: [
      {
        text: isEn ? 'Net' : '網絡',
        items: [
          { text: isEn ? 'Overview' : '概覽', link: `/${lang}/net/` }
        ]
      }
    ]
  }
}
