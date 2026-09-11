import { createInertiaApp } from '@inertiajs/svelte'
import { mount } from 'svelte'

createInertiaApp({
  progress: {
    delay: 80,
    color: "#d4a06a",
    includeCSS: true,
    showSpinner: false,
  },
  resolve: (name) => {
    const pages = import.meta.glob('./pages/**/*.svelte', { eager: true })
    const page = pages[`./pages/${name}.svelte`]
    if (!page) throw new Error(`Inertia page not found: ${name}`)
    return page
  },
  setup({ el, App, props }) {
    mount(App, { target: el, props })
  },
  // Match pkg/cais/csrf double-submit cookie (not Laravel XSRF defaults).
  // Production sets Secure cookies, which use the __Host- prefix (#173);
  // Vite's PROD flag tracks ENV=production, so pick the matching name.
  http: {
    xsrfCookieName: import.meta.env.PROD ? '__Host-cais_csrf' : 'cais_csrf',
    xsrfHeaderName: 'X-CSRF-Token',
  },
})
