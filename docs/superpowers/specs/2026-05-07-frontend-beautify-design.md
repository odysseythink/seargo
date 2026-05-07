# 前端美化设计文档

> 日期：2026-05-07
> 主题：SearGo 暗黑主题前端美化

## 目标

将当前简陋的搜索界面升级为现代暗黑风格的搜索体验。

## 设计系统

### 色彩
- 背景：`#0f0f0f`
- 卡片背景：`#1a1a1a`
- 卡片边框：`rgba(255,255,255,0.08)`
- 主文本：`#e5e5e5`
- 次要文本：`#9ca3af`
- 强调色（蓝）：`#3b82f6`
- 链接：`#60a5fa`
- 引擎标签背景：按引擎分配

### 引擎标签色彩
| 引擎 | 颜色 |
|------|------|
| Google | `#ea4335` |
| Bing | `#00809d` |
| DuckDuckGo | `#de5833` |
| Brave | `#fb542b` |
| Wikipedia | `#3366cc` |
| Yahoo | `#6001d2` |

### 布局
- 居中布局，max-width: 800px
- 搜索页：大搜索框居中，聚焦发光效果
- 结果页：卡片式结果列表，圆角 12px

### 交互
- 搜索按钮 loading 旋转动画
- 结果卡片淡入动画 (stagger)
- 输入框聚焦时蓝色 glow

## 技术方案

- CSS 框架：Tailwind CSS v4
- 字体：系统字体栈
- 动画：CSS transitions + Tailwind animate

## 文件变更

- `web/package.json` — 添加 tailwindcss
- `web/src/index.css` — 全局样式 + Tailwind 指令
- `web/src/pages/SearchPage.tsx` — 重写为暗黑主题
- `web/vite.config.ts` — 配置 Tailwind
