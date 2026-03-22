# Uplink Web App: Landing Page & Polish Implementation Plan (Plan 3e)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the landing page, implement empty state screens, add browser push notifications, set up Playwright browser tests, and write the README with self-hosting guide.

**Architecture:** Landing page is a standalone Blade/Inertia page (no auth required). Empty states are Vue components embedded in existing pages. Push notifications use the Web Push API with a service worker. Playwright tests cover critical user flows.

**Tech Stack:** Vue 3, Web Push API, Playwright, Chief brand assets from `assets/`

**Spec:** `docs/superpowers/specs/2026-03-22-uplink-webapp-design.md`

**Prerequisite:** Plans 3a-3d should be substantially complete. This plan adds polish and can be worked on incrementally.

---

## File Structure

```
resources/js/
├── Pages/
│   └── Welcome.vue                     ← Landing page
├── Components/
│   ├── EmptyState.vue                  ← reusable empty state component
│   ├── EmptyDevices.vue                ← "Connect your first Chief server"
│   ├── EmptyProjects.vue               ← "This workspace is empty"
│   ├── EmptyPrds.vue                   ← "Create your first PRD"
│   ├── EmptyServers.vue                ← "Launch a Chief server in the cloud"
│   └── EmptyTeamMembers.vue            ← "Invite your team"
public/
├── sw.js                               ← Service worker for push notifications
└── images/                             ← Screenshots, hero images
app/Http/Controllers/
└── WelcomeController.php               ← Landing page
tests/
├── Browser/                            ← Playwright tests
│   ├── playwright.config.js
│   ├── auth.spec.js
│   ├── prd-chat.spec.js
│   └── device.spec.js
README.md
docs/
└── self-hosting.md
```

---

### Task 1: Landing Page

**Files:**
- Create: `app/Http/Controllers/WelcomeController.php`
- Create: `resources/js/Pages/Welcome.vue`

- [ ] **Step 1: Create WelcomeController**

```php
// app/Http/Controllers/WelcomeController.php
<?php

namespace App\Http\Controllers;

use Inertia\Inertia;

class WelcomeController extends Controller
{
    public function __invoke()
    {
        if (auth()->check()) {
            return redirect('/');
        }

        return Inertia::render('Welcome');
    }
}
```

- [ ] **Step 2: Update route**

Change the root route for unauthenticated users:

```php
// routes/web.php
Route::get('/', function () {
    if (auth()->check()) {
        return app(DashboardController::class)->index();
    }
    return app(WelcomeController::class)();
})->name('home');
```

Or more cleanly, set up the middleware to redirect:
```php
Route::get('/', WelcomeController::class)->name('welcome')->middleware('guest');
Route::get('/dashboard', [DashboardController::class, 'index'])->name('dashboard')->middleware('auth');
```

- [ ] **Step 3: Create Welcome.vue**

```vue
<!-- resources/js/Pages/Welcome.vue -->
<script setup>
import { Link } from '@inertiajs/vue3';

defineOptions({ layout: null }); // no app layout for landing page
</script>

<template>
    <div class="min-h-screen bg-bg text-text">
        <!-- Nav -->
        <nav class="flex items-center justify-between px-6 py-4 max-w-5xl mx-auto">
            <div class="flex items-center gap-2">
                <img src="/images/logo.svg" alt="Chief" class="h-6" />
                <span class="text-text-heading font-semibold text-sm">Uplink</span>
            </div>
            <div class="flex items-center gap-3">
                <Link href="/login" class="text-sm text-text-secondary hover:text-text">Sign in</Link>
                <Link href="/register" class="text-sm bg-interactive text-bg px-4 py-1.5 rounded font-medium hover:opacity-90">Sign up</Link>
            </div>
        </nav>

        <!-- Hero -->
        <section class="text-center px-6 pt-20 pb-16 max-w-3xl mx-auto">
            <h1 class="text-4xl md:text-5xl font-bold text-text-heading tracking-tight leading-tight mb-4">
                Remote control for Chief
            </h1>
            <p class="text-lg text-text-secondary max-w-xl mx-auto mb-8">
                Monitor runs, create PRDs, review diffs, and manage your Chief servers — from anywhere.
                All AI execution happens on your machine using your own Claude subscription.
            </p>
            <div class="flex items-center justify-center gap-3">
                <Link href="/register" class="bg-interactive text-bg px-6 py-2.5 rounded font-medium text-sm hover:opacity-90">Get Started</Link>
                <a href="https://github.com/minicodemonkey/chief-uplink" class="text-sm text-text-secondary hover:text-text border border-border px-6 py-2.5 rounded">View on GitHub</a>
            </div>
        </section>

        <!-- Screenshot -->
        <section class="px-6 pb-20 max-w-4xl mx-auto">
            <div class="bg-bg-card border border-border rounded overflow-hidden">
                <img src="/images/screenshot-run.png" alt="Chief Uplink run monitoring" class="w-full" />
            </div>
        </section>

        <!-- How it works -->
        <section class="px-6 pb-20 max-w-3xl mx-auto">
            <h2 class="text-2xl font-semibold text-text-heading text-center mb-12">How it works</h2>
            <div class="grid md:grid-cols-3 gap-8">
                <div class="text-center">
                    <div class="text-3xl font-bold text-interactive mb-3">1</div>
                    <h3 class="font-medium text-text-heading mb-2">Install Chief</h3>
                    <p class="text-sm text-text-secondary">Install the Chief CLI on your machine or VPS.</p>
                </div>
                <div class="text-center">
                    <div class="text-3xl font-bold text-interactive mb-3">2</div>
                    <h3 class="font-medium text-text-heading mb-2">Connect</h3>
                    <p class="text-sm text-text-secondary">Run <code class="font-mono text-xs bg-bg-surface px-1.5 py-0.5 rounded">chief login</code> and <code class="font-mono text-xs bg-bg-surface px-1.5 py-0.5 rounded">chief serve</code></p>
                </div>
                <div class="text-center">
                    <div class="text-3xl font-bold text-interactive mb-3">3</div>
                    <h3 class="font-medium text-text-heading mb-2">Manage</h3>
                    <p class="text-sm text-text-secondary">Open Uplink in your browser — from your phone, laptop, anywhere.</p>
                </div>
            </div>
        </section>

        <!-- Features -->
        <section class="px-6 pb-20 max-w-4xl mx-auto">
            <h2 class="text-2xl font-semibold text-text-heading text-center mb-12">Features</h2>
            <div class="grid md:grid-cols-2 gap-4">
                <div class="bg-bg-card border border-border rounded p-4">
                    <h3 class="font-medium text-text-heading text-sm mb-1">PRD Chat</h3>
                    <p class="text-xs text-text-secondary">Create and refine PRDs through a conversational interface with Claude.</p>
                </div>
                <div class="bg-bg-card border border-border rounded p-4">
                    <h3 class="font-medium text-text-heading text-sm mb-1">Run Monitoring</h3>
                    <p class="text-xs text-text-secondary">Watch Ralph loops in real-time. See story progress, Claude output, and tool usage.</p>
                </div>
                <div class="bg-bg-card border border-border rounded p-4">
                    <h3 class="font-medium text-text-heading text-sm mb-1">Code Review</h3>
                    <p class="text-xs text-text-secondary">GitHub-style syntax-highlighted diffs of what Claude built.</p>
                </div>
                <div class="bg-bg-card border border-border rounded p-4">
                    <h3 class="font-medium text-text-heading text-sm mb-1">Cloud Servers</h3>
                    <p class="text-xs text-text-secondary">One-click VPS provisioning on Hetzner or DigitalOcean. Chief pre-installed.</p>
                </div>
                <div class="bg-bg-card border border-border rounded p-4">
                    <h3 class="font-medium text-text-heading text-sm mb-1">Teams</h3>
                    <p class="text-xs text-text-secondary">Share device access with your team. Owner and Member roles.</p>
                </div>
                <div class="bg-bg-card border border-border rounded p-4">
                    <h3 class="font-medium text-text-heading text-sm mb-1">Open Source</h3>
                    <p class="text-xs text-text-secondary">Self-host on your own infrastructure. Free to use.</p>
                </div>
            </div>
        </section>

        <!-- Self-hosting note -->
        <section class="px-6 pb-20 max-w-3xl mx-auto text-center">
            <p class="text-sm text-text-secondary">
                Chief Uplink is open source and free. Use the hosted version at uplink.chiefloop.com
                or <a href="https://github.com/minicodemonkey/chief-uplink#self-hosting" class="text-text underline underline-offset-2">self-host it</a>.
            </p>
        </section>

        <!-- Footer -->
        <footer class="border-t border-border px-6 py-6 text-center">
            <p class="text-xs text-text-muted">Chief Uplink — built by <a href="https://github.com/minicodemonkey" class="text-text-secondary hover:text-text">MiniCodeMonkey</a></p>
        </footer>
    </div>
</template>
```

- [ ] **Step 4: Copy brand assets**

```bash
cp /Users/codemonkey/projects/chief/assets/logo.svg public/images/logo.svg
cp /Users/codemonkey/projects/chief/assets/mark.svg public/images/mark.svg
```

- [ ] **Step 5: Commit**

```bash
git add app/Http/Controllers/WelcomeController.php resources/js/Pages/Welcome.vue public/images/ routes/web.php
git commit -m "feat: add landing page"
```

---

### Task 2: Empty State Components

**Files:**
- Create: `resources/js/Components/EmptyState.vue`
- Create: 5 specific empty state components

- [ ] **Step 1: Create reusable EmptyState base**

```vue
<!-- resources/js/Components/EmptyState.vue -->
<script setup>
defineProps({
    title: String,
    description: String,
});
</script>

<template>
    <div class="flex flex-col items-center justify-center py-16 px-4">
        <h3 class="text-base font-medium text-text-heading mb-2">{{ title }}</h3>
        <p class="text-sm text-text-secondary text-center max-w-md mb-6">{{ description }}</p>
        <slot />
    </div>
</template>
```

- [ ] **Step 2: Create EmptyDevices**

```vue
<!-- resources/js/Components/EmptyDevices.vue -->
<script setup>
import { Link } from '@inertiajs/vue3';
import EmptyState from './EmptyState.vue';
</script>

<template>
    <EmptyState
        title="Connect your first Chief server"
        description="Chief runs on your machine — Uplink lets you control it from here."
    >
        <div class="flex gap-3">
            <div class="bg-bg-card border border-border rounded p-4 max-w-xs">
                <h4 class="text-sm font-medium text-text-heading mb-2">Run locally</h4>
                <code class="block text-xs font-mono bg-bg-surface rounded p-2 text-text-secondary mb-1">chief login</code>
                <code class="block text-xs font-mono bg-bg-surface rounded p-2 text-text-secondary">chief serve</code>
            </div>
            <Link href="/servers/create" class="bg-bg-card border border-border rounded p-4 max-w-xs hover:border-interactive/30 transition-colors">
                <h4 class="text-sm font-medium text-text-heading mb-2">Launch cloud server</h4>
                <p class="text-xs text-text-secondary">One-click VPS on Hetzner or DigitalOcean</p>
            </Link>
        </div>
    </EmptyState>
</template>
```

- [ ] **Step 3: Create remaining empty states**

Follow the same pattern for EmptyProjects, EmptyPrds, EmptyServers, EmptyTeamMembers. Each has a title, description, and actionable buttons/links. See spec Section "Empty States" for the exact content of each.

- [ ] **Step 4: Integrate into existing pages**

Add empty state checks to Devices/Index.vue, Devices/Show.vue, Projects/Show.vue, Servers/Index.vue, and Settings/Team.vue. Show the empty state component when the corresponding list is empty.

- [ ] **Step 5: Commit**

```bash
git add resources/js/Components/Empty*
git commit -m "feat: add empty state components for onboarding UX"
```

---

### Task 3: Browser Push Notifications

**Files:**
- Create: `public/sw.js`
- Create: `resources/js/composables/usePushNotifications.js`

- [ ] **Step 1: Create service worker**

```js
// public/sw.js
self.addEventListener('push', (event) => {
    const data = event.data?.json() ?? {};

    event.waitUntil(
        self.registration.showNotification(data.title ?? 'Chief Uplink', {
            body: data.body ?? 'A run has completed.',
            icon: '/images/mark.svg',
            badge: '/images/mark.svg',
            data: { url: data.url ?? '/' },
        })
    );
});

self.addEventListener('notificationclick', (event) => {
    event.notification.close();
    event.waitUntil(
        clients.openWindow(event.notification.data.url)
    );
});
```

- [ ] **Step 2: Create push notification composable**

```js
// resources/js/composables/usePushNotifications.js
import { ref } from 'vue';

export function usePushNotifications() {
    const isSupported = ref('serviceWorker' in navigator && 'PushManager' in window);
    const isSubscribed = ref(false);

    async function subscribe() {
        if (!isSupported.value) return;

        const registration = await navigator.serviceWorker.register('/sw.js');
        const subscription = await registration.pushManager.subscribe({
            userVisibleOnly: true,
            applicationServerKey: import.meta.env.VITE_VAPID_PUBLIC_KEY,
        });

        // Send subscription to server
        await fetch('/api/push-subscriptions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(subscription),
        });

        isSubscribed.value = true;
    }

    return { isSupported, isSubscribed, subscribe };
}
```

- [ ] **Step 3: Add server-side push notification on run completion**

In the `StateHandler`, when processing `state.run.completed`, trigger a web push notification to team members who are not currently connected via Reverb. This requires a `push_subscriptions` table and the `web-push` PHP library.

```bash
composer require minishlink/web-push
```

- [ ] **Step 4: Commit**

```bash
git add public/sw.js resources/js/composables/usePushNotifications.js
git commit -m "feat: add browser push notifications for run completion"
```

---

### Task 4: Playwright Browser Tests

**Files:**
- Create: `tests/Browser/playwright.config.js`
- Create: `tests/Browser/auth.spec.js`
- Create: `tests/Browser/prd-chat.spec.js`

- [ ] **Step 1: Install Playwright**

```bash
npm install -D @playwright/test
npx playwright install chromium
```

- [ ] **Step 2: Create Playwright config**

```js
// tests/Browser/playwright.config.js
import { defineConfig } from '@playwright/test';

export default defineConfig({
    testDir: '.',
    baseURL: 'http://localhost:8000',
    use: {
        viewport: { width: 375, height: 812 }, // iPhone-sized (mobile-first)
        screenshot: 'only-on-failure',
    },
    webServer: {
        command: 'php artisan serve --port=8000',
        port: 8000,
        reuseExistingServer: true,
    },
});
```

- [ ] **Step 3: Create auth test**

```js
// tests/Browser/auth.spec.js
import { test, expect } from '@playwright/test';

test('can register and see dashboard', async ({ page }) => {
    await page.goto('/register');

    await page.fill('input[placeholder="Name"]', 'Test User');
    await page.fill('input[placeholder="Email"]', `test-${Date.now()}@example.com`);
    await page.fill('input[placeholder="Password"]', 'password123');
    await page.fill('input[placeholder="Confirm Password"]', 'password123');
    await page.click('button[type="submit"]');

    await expect(page).toHaveURL('/');
    await expect(page.locator('text=Welcome to Chief Uplink')).toBeVisible();
});

test('can login with email', async ({ page }) => {
    // Assumes a seeded user exists
    await page.goto('/login');
    await expect(page.locator('text=Sign in to your account')).toBeVisible();
    await expect(page.locator('text=Sign in with GitHub')).toBeVisible();
});

test('shows landing page when not logged in', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('text=Remote control for Chief')).toBeVisible();
});
```

- [ ] **Step 4: Create PRD chat test (mobile viewport)**

```js
// tests/Browser/prd-chat.spec.js
import { test, expect } from '@playwright/test';

test('empty state shows on devices page', async ({ page }) => {
    // Login first (helper needed)
    await page.goto('/login');
    await page.fill('input[placeholder="Email"]', 'test@example.com');
    await page.fill('input[placeholder="Password"]', 'password123');
    await page.click('button[type="submit"]');

    await page.goto('/devices');
    await expect(page.locator('text=Connect your first Chief server')).toBeVisible();
});
```

- [ ] **Step 5: Add to CI**

Update `.github/workflows/ci.yml` to include Playwright:

```yaml
  browser-tests:
    runs-on: ubuntu-latest
    needs: tests
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 22 }
      - run: npm ci
      - run: npx playwright install --with-deps chromium
      - run: npm run build
      - run: npx playwright test --config=tests/Browser/playwright.config.js
```

- [ ] **Step 6: Commit**

```bash
git add tests/Browser/ .github/workflows/ci.yml
git commit -m "feat: add Playwright browser tests for auth and empty states"
```

---

### Task 5: README

**Files:**
- Create: `README.md`

- [ ] **Step 1: Write README**

```markdown
# Chief Uplink

Remote management and control for [Chief](https://chiefloop.com) — monitor runs, create PRDs, review diffs, and manage servers from anywhere.

[screenshot placeholder]

## What is Chief Uplink?

Chief Uplink is a web companion for Chief that lets you remote-control your running `chief serve` instances from any browser. The web app handles accounts, device authentication, and real-time relay between your browser and your Chief server. It does zero AI work — all Claude Code execution happens on your machine or VPS using your own Claude subscription.

Chief connects outbound to the web app (like Plex or Tailscale), so there's no port forwarding or firewall configuration needed.

## Features

- **PRD Chat** — create and refine PRDs through a conversational interface with Claude
- **Run Monitoring** — watch Ralph loops in real-time with story progress and live Claude output
- **Code Review** — syntax-highlighted diffs of what Claude built
- **File Browser** — browse and view project files remotely
- **Cloud Servers** — one-click VPS provisioning on Hetzner or DigitalOcean
- **Teams** — share device access with your team
- **Push Notifications** — get notified when overnight runs finish

## Quick Start (Hosted)

1. Sign up at [uplink.chiefloop.com](https://uplink.chiefloop.com)
2. Install Chief: see [chiefloop.com/guide/installation](https://chiefloop.com/guide/installation)
3. Connect: `chief login` → approve in browser → `chief serve`
4. Open Uplink in your browser

## Development Setup

### Prerequisites

- Docker and Docker Compose
- Node.js 22+

### Setup

```bash
git clone https://github.com/minicodemonkey/chief-uplink.git
cd chief-uplink
cp .env.example .env
docker compose up -d
docker compose exec app php artisan key:generate
docker compose exec app php artisan migrate
npm install
npm run dev
```

Visit http://localhost:8000

### Running Tests

```bash
# PHP tests
docker compose exec app php artisan test

# Vue component tests
npx vitest run

# Browser tests
npx playwright test --config=tests/Browser/playwright.config.js

# Code style
docker compose exec app ./vendor/bin/pint --test
```

## Self-Hosting

See [docs/self-hosting.md](docs/self-hosting.md) for complete self-hosting instructions.

### Quick Self-Host

```bash
git clone https://github.com/minicodemonkey/chief-uplink.git
cd chief-uplink
cp .env.example .env
# Edit .env: set APP_URL, APP_KEY, database credentials
docker compose -f docker-compose.prod.yml up -d
docker compose exec app php artisan migrate
```

## Tech Stack

- Laravel 13 / Octane (FrankenPHP)
- Vue 3 / Inertia.js / Tailwind 4
- Laravel Reverb (real-time)
- MariaDB
- Pest / Vitest / Playwright

## License

MIT
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add README with quick start and development setup"
```

---

### Task 6: Self-Hosting Guide

**Files:**
- Create: `docs/self-hosting.md`

- [ ] **Step 1: Write self-hosting guide**

Cover: requirements (Docker, domain, SSL), deployment steps, environment variables (APP_KEY critical for encryption), reverse proxy config (Caddy recommended, with WebSocket upgrade for `/ws/device` and Reverb), database setup, Redis, backups, and updates. See spec "Self-Hosting Guide" section for full requirements.

- [ ] **Step 2: Commit**

```bash
git add docs/self-hosting.md
git commit -m "docs: add self-hosting guide"
```
