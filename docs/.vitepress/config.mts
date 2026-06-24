import { defineConfig } from 'vitepress'

export default defineConfig({
    base: '/hosomaki/',
    appearance: 'force-light',
    title: 'Hosomaki',
    description: 'Linux diagnostics in plain language. No cloud. No telemetry. Your system, your data.',

    lang: 'en-US',
    lastUpdated: true,
    cleanUrls: false,

    head: [
        ['link', { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' }],
        ['meta', { name: 'theme-color', content: '#0f1117' }],
        ['meta', { property: 'og:type', content: 'website' }],
        ['meta', { property: 'og:title', content: 'Hosomaki — Linux diagnostics in plain language' }],
        ['meta', { property: 'og:description', content: 'No cloud. No telemetry. Your system, your data.' }],
    ],

    themeConfig: {
        logo: '/hosomaki_logo.png',
        siteTitle: 'Hosomaki',

        nav: [
            { text: 'Docs', link: '/guide/introduction', activeMatch: '/guide/' },
            { text: 'Reference', link: '/reference/commands', activeMatch: '/reference/' },
        ],

        sidebar: [
            {
                text: 'Getting Started',
                collapsed: false,
                items: [
                    { text: 'Introduction', link: '/guide/introduction' },
                    { text: 'Installation', link: '/guide/installation' },
                    { text: 'Quick Start', link: '/guide/quickstart' },
                    { text: 'Configuration', link: '/guide/configuration' },
                    { text: 'Troubleshooting', link: '/guide/troubleshooting' },
                ],
            },
            {
                text: 'Core CLI',
                collapsed: false,
                items: [
                    { text: 'explain', link: '/reference/explain' },
                    { text: 'status', link: '/reference/status' },
                    { text: 'doctor', link: '/reference/doctor' },
                    { text: 'audit', link: '/reference/audit' },
                    { text: 'watch', link: '/reference/watch' },
                    { text: 'why', link: '/reference/why' },
                    { text: 'ports', link: '/reference/ports' },
                    { text: 'timers', link: '/reference/timers' },
                    { text: 'crons', link: '/reference/crons' },
                    { text: 'mounts', link: '/reference/mounts' },
                    { text: 'updates', link: '/reference/updates' },
                    { text: 'history', link: '/reference/history' },
                    { text: 'shell-integration', link: '/reference/shell-integration' },
                ],
            },
            {
                text: 'Architecture',
                collapsed: false,
                items: [
                    { text: 'Overview', link: '/guide/architecture' },
                    { text: 'AI Pipeline', link: '/guide/pipeline' },
                    { text: 'Sanitisation', link: '/guide/sanitisation' },
                    { text: 'Data Privacy', link: '/guide/privacy' },
                ],
            },
            {
                text: 'Daemon Layer',
                collapsed: true,
                items: [
                    { text: 'Overview', link: '/daemon/overview' },
                    { text: 'Configuration', link: '/daemon/configuration' },
                ],
            },
            {
                text: 'Future: Native UI',
                collapsed: true,
                items: [
                    { text: 'Roadmap', link: '/ui/roadmap' },
                ],
            },
        ],

        socialLinks: [
            { icon: 'github', link: 'https://github.com/rivernova/hosomaki' },
        ],

        search: {
            provider: 'local',
            options: {
                detailedView: true,
            },
        },

        footer: {
            message: 'Released under the <a href="https://mozilla.org/en-US/MPL/2.0/">Mozilla Public License 2.0</a>.',
            copyright: 'Copyright © 2026–present rivernova',
        },

        editLink: {
            pattern: 'https://github.com/rivernova/hosomaki/edit/main/docs/:path',
            text: 'Edit this page on GitHub',
        },

        outline: {
            level: [2, 3],
            label: 'On this page',
        },

        docFooter: {
            prev: 'Previous',
            next: 'Next',
        },

        returnToTopLabel: 'Back to top',
        sidebarMenuLabel: 'Menu',
        darkModeSwitchLabel: 'Appearance',
    },
})
