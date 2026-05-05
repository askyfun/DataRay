# Cerebrum

> OpenWolf's learning memory. Updated automatically as the AI learns from interactions.
> Do not edit manually unless correcting an error.
> Last updated: 2026-05-05

## User Preferences

<!-- How the user likes things done. Code style, tools, patterns, communication. -->

## Key Learnings

- **Project:** data-insights
- **Description:** 拖拽式 BI 可视化分析平台 MVP

- **[2026-05-05] pnpm 严格依赖解析**: 从 npm 切换到 pnpm 后，antd 的传递依赖 `@ant-design/icons` 和 `dayjs` 必须显式安装。npm 会自动 hoist 这些依赖，pnpm 不会。这是 pnpm 的正确行为——依赖应该显式声明。

## Do-Not-Repeat

<!-- Mistakes made and corrected. Each entry prevents the same mistake recurring. -->
<!-- Format: [YYYY-MM-DD] Description of what went wrong and what to do instead. -->

## Decision Log

<!-- Significant technical decisions with rationale. Why X was chosen over Y. -->
