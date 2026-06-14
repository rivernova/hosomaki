<template>
  <div class="lp">

    <SiteNav />

    <!-- ── HERO ─────────────────────────────────────────────────────── -->
    <section class="hero">
      <span class="eyebrow">hosomaki awake</span>
      <h1 class="hero-h1">
        A lightweight<br class="br-md" />
        <em class="c-salmon">intelligence</em> layer<br class="br-md" />
        for <em class="c-teal">Linux</em>
      </h1>
      <p class="hero-sub">
        Reads your system and tells you what's happening in plain language.
        No cloud. No telemetry.
      </p>
      <div class="hero-btns">
        <a class="btn btn-primary" href="/guide/installation">Get started</a>
        <a class="btn btn-ghost"  href="/guide/introduction">Read the docs</a>
        <a class="btn btn-ghost"  href="https://github.com/rivernova/hosomaki" target="_blank" rel="noopener">GitHub</a>
      </div>
      <div class="hero-tags">
        <span v-for="tag in heroTags" :key="tag">{{ tag }}</span>
      </div>
    </section>

    <div class="divider-wrap"><hr class="divider" /></div>

    <!-- ── COMMANDS ──────────────────────────────────────────────────── -->
    <section class="section reveal" aria-labelledby="commands-heading">
      <div class="section-header">
        <div>
          <span class="label">commands</span>
          <h2 id="commands-heading" class="section-h2">
            Everything your system<br />needs to say.
          </h2>
        </div>
        <p class="section-lead">
          Eleven commands. One pipeline. Each one sanitised before
          prompting and interpreted entirely on your machine.
        </p>
      </div>
      <div class="cmd-grid">
        <a
          v-for="cmd in commands"
          :key="cmd.slug"
          class="cmd-card"
          :href="`/reference/${cmd.slug}`"
        >
          <span class="cmd-mono">{{ cmd.name }}</span>
          <span class="cmd-title">{{ cmd.title }}</span>
          <span class="cmd-desc" v-html="cmd.desc" />
          <span class="cmd-arrow" aria-hidden="true">Reference →</span>
        </a>
      </div>
    </section>

    <div class="divider-wrap"><hr class="divider" /></div>

    <!-- ── PIPELINE ──────────────────────────────────────────────────── -->
    <section class="section reveal" aria-labelledby="pipeline-heading">
      <div class="pipe-wrap">
        <div class="pipe-left">
          <span class="label">architecture</span>
          <h2 id="pipeline-heading" class="section-h2">
            One pipeline.
          </h2>
          <p class="section-lead pipe-lead">
            Sanitisation is the mandatory first stage of
            every invocation.
            Nothing raw reaches the model.
          </p>
          <ol class="pipe-steps" aria-label="Pipeline stages">
            <li v-for="step in pipelineSteps" :key="step.num" class="pipe-step">
              <span class="pipe-num" aria-hidden="true">{{ step.num }}</span>
              <div>
                <p class="pipe-name">{{ step.name }}</p>
                <p class="pipe-desc">{{ step.desc }}</p>
              </div>
            </li>
          </ol>
        </div>
        <div class="pipe-right" aria-hidden="true">
          <div class="code-block">
            <span class="c-lbl">// every command, no exceptions</span>
            <span class="c-txt">collect</span><br />
            <span class="c-arr">  ↓</span><br />
            <span class="c-hl">sanitise</span><span class="c-dim"></span><br />
            <span class="c-arr">  ↓</span><br />
            <span class="c-txt">prompt</span><br />
            <span class="c-arr">  ↓</span><br />
            <span class="c-txt">validate</span><br />
            <span class="c-arr">  ↓</span><span class="c-dim">   // failed?</span><br />
            <span class="c-txt">repair</span><span class="c-dim">      // with context</span><br />
            <span class="c-arr">  ↓</span><br />
            <span class="c-ok">render</span>
          </div>
        </div>
      </div>
    </section>

    <!-- ── PRIVACY — dark band ────────────────────────────────────────── -->
    <section class="dark-band reveal" aria-labelledby="privacy-heading">
      <div class="dark-inner">
        <div>
          <span class="label dark-label">privacy</span>
          <h2 id="privacy-heading" class="section-h2 dark-h2">
            Stays on<br />your machine.
          </h2>
          <p class="dark-lead">
            The model runs locally. No data leaves your system.
          </p>
        </div>
        <ul class="priv-list" aria-label="Privacy guarantees">
          <li v-for="item in privacyItems" :key="item.title" class="priv-item">
            <span class="priv-dot" :class="{ 'priv-dot--accent': item.accent }" aria-hidden="true" />
            <div>
              <p class="priv-title">{{ item.title }}</p>
              <p class="priv-desc">{{ item.desc }}</p>
            </div>
          </li>
        </ul>
      </div>
    </section>

   <!-- ── STORY ──────────────────────────────────────────────────────── -->
   <section class="section reveal" aria-labelledby="story-heading">
     <div class="story-wrap">
       <div class="story-left">
         <span class="label">why Hosomaki exists</span>
         <h2 id="story-heading" class="section-h2 story-h2">
           Your system has always known what’s happening. Hosomaki just grew into the layer that reveals it.
         </h2>
       </div>
       <div class="story-right">
         <p>
           Hosomaki began as a simple idea. It started as a tool that could translate logs into something
           mortals like you and me could read without digging through journal entries or kernel messages.
           But the deeper the project went, the clearer it became that translation wasn’t
           the real goal. Linux already speaks, constantly. The challenge is making
           sense of everything it says.
         </p>
         <p>
           That’s where Hosomaki evolved. It no longer just interprets logs. It builds a
           structured understanding of your system by correlating signals, validating
           assumptions, and explaining what matters and why. It reads the same sources you
           would, strips anything sensitive, and produces a grounded analysis. Not a
           chatbot reply, but a deterministic pipeline with schema checks and repair loops
           before anything reaches your screen.
         </p>
         <p>
           The model runs on your machine. The data stays on your machine.
           Your data belongs to you.
         </p>
       </div>
     </div>
   </section>

    <div class="divider-wrap"><hr class="divider" /></div>

    <!-- ── INSTALL — dark band ────────────────────────────────────────── -->
    <section class="install-band reveal" aria-labelledby="install-heading">
      <div class="install-inner">

        <div class="install-left">
          <span class="label dark-label">open source · mpl-2.0</span>
          <h2 id="install-heading" class="section-h2 dark-h2">
            Clone and run.
          </h2>
          <p class="dark-lead" style="margin-bottom: 2rem">
            No account. No config file on first run.
            Requires Go 1.23+ and Ollama running locally.
          </p>
          <dl class="install-manifest">
            <div v-for="item in manifest" :key="item.key" class="manifest-row">
              <dt class="manifest-key">{{ item.key }}</dt>
              <dd class="manifest-val">{{ item.val }}</dd>
            </div>
          </dl>
        </div>

        <div class="install-right">
          <ol class="install-steps" aria-label="Installation steps">
            <li
              v-for="(step, i) in installSteps"
              :key="i"
              class="install-step"
            >
              <span class="is-num" aria-hidden="true">{{ String(i + 1).padStart(2, '0') }}</span>
              <div class="is-body">
                <span class="is-comment">{{ step.comment }}</span>
                <button
                  class="is-cmd"
                  :class="{ 'is-cmd--copied': copiedIdx === i }"
                  type="button"
                  :aria-label="`Copy command: ${step.cmd}`"
                  @click="copyStep(i, step.cmd)"
                >
                  <span class="is-prompt" aria-hidden="true">$</span>
                  <span class="is-text">{{ step.cmd }}</span>
                  <span class="is-hint" aria-live="polite">
                    {{ copiedIdx === i ? '✓ copied' : 'copy' }}
                  </span>
                </button>
              </div>
            </li>
          </ol>
          <div class="install-actions">
            <a class="btn btn-install-primary" href="/guide/installation">Full install guide</a>
            <a class="btn btn-install-ghost"   href="https://github.com/rivernova/hosomaki" target="_blank" rel="noopener">
              GitHub →
            </a>
          </div>
        </div>

      </div>
    </section>

    <!-- ── FOOTER ─────────────────────────────────────────────────────── -->
    <footer class="site-footer">
      <span class="footer-copy">
        © 2026–present rivernova ·
        <a href="https://mozilla.org/en-US/MPL/2.0/" target="_blank" rel="noopener">MPL-2.0</a>
        · <a href="https://github.com/rivernova/hosomaki" target="_blank" rel="noopener">GitHub</a>
      </span>
      <nav class="footer-links" aria-label="Footer navigation">
        <a href="/guide/introduction">Docs</a>
        <a href="/reference/commands">Reference</a>
        <a href="/guide/architecture">Architecture</a>
        <a href="/guide/privacy">Privacy</a>
      </nav>
    </footer>

  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import SiteNav from './SiteNav.vue'

// ─── hero ──────────────────────────────────────────────────────────────────
const heroTags = ['No cloud', 'No telemetry', 'Local · Ollama', 'Read-only', 'MPL-2.0']

// ─── commands ──────────────────────────────────────────────────────────────
const commands = [
  {
    slug: 'explain',
    name: 'hosomaki explain',
    title: 'Explain anything',
    desc: 'Services, boot logs, dmesg, piped streams, inline messages, running processes. If it logs, explain reads it.',
  },
  {
    slug: 'status',
    name: 'hosomaki status',
    title: 'System health',
    desc: 'A snapshot of uptime, memory, disk, failed services, and recent errors.',
  },
  {
    slug: 'doctor',
    name: 'hosomaki doctor',
    title: 'Full diagnosis',
    desc: 'Deeper than status. Surfaces concrete investigation steps for every identified issue.',
  },
  {
    slug: 'watch',
    name: 'hosomaki watch',
    title: 'Real-time tail',
    desc: 'Tails a service journal and explains error bursts as they arrive.',
  },
  {
    slug: 'audit',
    name: 'hosomaki audit',
    title: 'Change detection',
    desc: 'Snapshots your system state and diffs against it. Files, packages, listening ports, and unit states.',
  },
  {
    slug: 'ports',
    name: 'hosomaki ports',
    title: 'Port exposure',
    desc: 'Lists every listening socket with process names and flags anything unexpected or worth investigating.',
  },
  {
    slug: 'timers',
    name: 'hosomaki timers',
    title: 'Systemd timers',
    desc: 'Inspects all timers and flags any that have failed.',
  },
  {
    slug: 'crons',
    name: 'hosomaki crons',
    title: 'Cron jobs',
    desc: 'Reads all crontab files and explains each job. Flags broken or suspicious entries.',
  },
  {
    slug: 'shell-integration',
    name: 'shell-integration',
    title: 'Auto-explain failures',
    desc: 'Prefix any command with <code class="ic">explain</code> and failed commands get diagnosed automatically.',
  },
]

// ─── pipeline ──────────────────────────────────────────────────────────────
const pipelineSteps = [
  {
    num: '01',
    name: 'Collect',
    desc: 'Raw system data. Never stored, never modified.',
  },
  {
    num: '02',
    name: 'Sanitise',
    desc: 'IPs, hostnames, paths, UUIDs, and credentials stripped before the model sees anything.',
  },
  {
    num: '03',
    name: 'Prompt',
    desc: 'A tightly constrained prompt with a strict JSON schema.',
  },
  {
    num: '04',
    name: 'Validate & Repair',
    desc: 'Schema and semantic checks on every response.',
  },
  {
    num: '05',
    name: 'Render',
    desc: 'Typed, validated output formatted for the terminal. Raw model text never reaches the screen.',
  },
]

// ─── privacy ───────────────────────────────────────────────────────────────
const privacyItems = [
  {
    title: 'Local model via Ollama',
    desc: 'All inference runs on localhost. Works with any model Ollama supports.',
    accent: false,
  },
  {
    title: 'Sanitisation before prompting',
    desc: 'IPs, hostnames, paths, UUIDs, and credentials stripped before the model sees anything.',
    accent: false,
  },
  {
    title: 'Zero telemetry',
    desc: 'No analytics, no crash reports, no usage metrics. The binary makes no outbound connections except to your Ollama instance.',
    accent: false,
  },
  {
    title: 'Strictly read-only',
    desc: 'Every command collects data and surfaces insights. None of them modify the system.',
    accent: true,
  },
]

// ─── install ───────────────────────────────────────────────────────────────
const manifest = [
  { key: 'license',     val: 'Mozilla Public License 2.0' },
  { key: 'runtime',     val: 'Linux · Go 1.23+' },
  { key: 'AI backend',  val: 'Ollama (local, localhost)' },
  { key: 'telemetry',   val: 'none, ever' },
]

// Each entry is exactly one shell command — no chaining with &&
const installSteps = [
  {
    comment: '# install Ollama',
    cmd: 'curl -fsSL https://ollama.com/install.sh | sh',
  },
  {
    comment: '# pull a model (llama3.1 is the default)',
    cmd: 'ollama pull llama3.1',
  },
  {
    comment: '# clone the repo',
    cmd: 'git clone https://github.com/rivernova/hosomaki.git && cd hosomaki',
  },
  {
    comment: '# build and install',
    cmd: 'sudo make install',
  },
]

const copiedIdx = ref(-1)
let copyTimer = null

function copyStep(idx, cmd) {
  if (!navigator.clipboard) return
  navigator.clipboard.writeText(cmd).then(() => {
    copiedIdx.value = idx
    clearTimeout(copyTimer)
    copyTimer = setTimeout(() => { copiedIdx.value = -1 }, 2000)
  })
}

// ─── scroll reveal ─────────────────────────────────────────────────────────
onMounted(() => {
  if (!('IntersectionObserver' in window)) {
    document.querySelectorAll('.reveal').forEach(el => el.classList.add('in'))
    return
  }
  const io = new IntersectionObserver(
    (entries) => {
      entries.forEach(e => {
        if (e.isIntersecting) {
          e.target.classList.add('in')
          io.unobserve(e.target)
        }
      })
    },
    { threshold: 0.05, rootMargin: '0px 0px -30px 0px' },
  )
  document.querySelectorAll('.reveal').forEach(el => io.observe(el))
})
</script>

<style>
/* ── Design tokens ──────────────────────────────────────────────────── */

.lp {
  --bg:     #f2efe9;
  --bg-alt: #ebe8e2;
  --ink:    #2e2b27;
  --ink-2:  #6b6660;
  --ink-3:  #a09b95;
  --salmon: #e06b50;
  --teal:   #5aa89a;
  --border: rgba(46, 43, 39, .1);
  --nav-h:  60px;
  --pad:    clamp(1.25rem, 5vw, 3rem);
  --max:    1100px;
  --mono:   'JetBrains Mono', 'Fira Code', monospace;

  font-family: var(--mono);
  background: var(--bg);
  color: var(--ink);
  -webkit-font-smoothing: antialiased;
  min-height: 100vh;
  overflow-x: hidden;
}

/* ── Shared atoms ───────────────────────────────────────────────────── */
.label {
  font-family: var(--mono);
  font-size: .66rem;
  color: var(--ink-3);
  letter-spacing: .12em;
  text-transform: uppercase;
  display: block;
  margin-bottom: 1.1rem;
}

.eyebrow {
  font-family: var(--mono);
  font-size: .7rem;
  color: var(--teal);
  letter-spacing: .12em;
  text-transform: uppercase;
  display: block;
  margin-bottom: 2rem;
}
.eyebrow::before { content: '> '; color: var(--ink-3); }

.section-h2 {
  font-size: clamp(1.55rem, 4vw, 2.6rem);
  font-weight: 700;
  letter-spacing: -.034em;
  line-height: 1.1;
  color: var(--ink);
  margin-bottom: .85rem;
}

.section-lead {
  font-size: .95rem;
  line-height: 1.78;
  color: var(--ink-2);
}

.btn {
  display: inline-flex;
  align-items: center;
  text-decoration: none;
  border-radius: 7px;
  font-size: .86rem;
  font-weight: 500;
  padding: .56rem 1.2rem;
  white-space: nowrap;
  transition: background .18s, color .18s, border-color .18s;
}
.btn-primary {
  background: var(--ink);
  color: var(--bg);
  border: 1px solid var(--ink);
}
.btn-primary:hover { background: var(--salmon); border-color: var(--salmon); }
.btn-ghost {
  background: transparent;
  color: var(--ink);
  border: 1px solid var(--border);
}
.btn-ghost:hover { border-color: var(--ink-3); }

.divider-wrap { padding: 0 var(--pad); max-width: var(--max); margin: 0 auto; }
.divider { border: none; border-top: 1px solid var(--border); }

.section {
  padding: clamp(4rem, 8vw, 7.5rem) var(--pad);
  max-width: var(--max);
  margin: 0 auto;
}

/* ── Hero ───────────────────────────────────────────────────────────── */
.hero {
  min-height: 100svh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: calc(var(--nav-h) + clamp(3rem, 6vw, 5rem)) var(--pad) clamp(3rem, 5vw, 5rem);
}

.hero-h1 {
  font-size: clamp(2.5rem, 9vw, 6.4rem);
  font-weight: 700;
  letter-spacing: -.045em;
  line-height: .97;
  color: var(--ink);
  margin-bottom: 1.6rem;
  max-width: 840px;
}
.hero-h1 em   { font-style: normal; }
.c-salmon     { color: var(--salmon); }
.c-teal       { color: var(--teal); }

.hero-sub {
  font-size: clamp(.92rem, 2vw, 1.1rem);
  line-height: 1.75;
  color: var(--ink-2);
  max-width: 500px;
  margin: 0 auto 2.4rem;
}

.hero-btns {
  display: flex;
  align-items: center;
  gap: .65rem;
  flex-wrap: wrap;
  justify-content: center;
  margin-bottom: 2.75rem;
}

.hero-tags {
  display: flex;
  align-items: center;
  gap: clamp(.6rem, 2.5vw, 1.6rem);
  flex-wrap: wrap;
  justify-content: center;
}
.hero-tags span {
  font-family: var(--mono);
  font-size: .69rem;
  color: var(--ink-3);
  letter-spacing: .04em;
}

/* ── Commands grid ──────────────────────────────────────────────────── */
.section-header {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  align-items: end;
  margin-bottom: 3rem;
}
.section-header .section-lead { max-width: 400px; }

.cmd-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  border: 1px solid var(--border);
  border-radius: 14px;
  overflow: hidden;
}

.cmd-card {
  display: flex;
  flex-direction: column;
  gap: .42rem;
  padding: 1.65rem 1.45rem;
  border-right: 1px solid var(--border);
  border-bottom: 1px solid var(--border);
  text-decoration: none;
  color: inherit;
  background: var(--bg);
  transition: background .2s;
}
.cmd-card:nth-child(3n)        { border-right: none; }
.cmd-card:nth-last-child(-n+3) { border-bottom: none; }
.cmd-card:hover                { background: var(--bg-alt, #ebe8e2); }
.cmd-card:hover .cmd-arrow     { color: var(--salmon); }

.cmd-mono  { font-family: var(--mono); font-size: .69rem; color: var(--salmon); letter-spacing: .04em; margin-bottom: .18rem; }
.cmd-title { font-size: .94rem; font-weight: 600; letter-spacing: -.012em; color: var(--ink); }
.cmd-desc  { font-size: .82rem; line-height: 1.65; color: var(--ink-2); }
.cmd-arrow { font-family: var(--mono); font-size: .69rem; color: var(--teal); margin-top: auto; padding-top: .6rem; transition: color .15s; }

/* ── Pipeline ───────────────────────────────────────────────────────── */
.pipe-wrap {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 5rem;
  align-items: start;
}
.pipe-lead  { margin-bottom: 2.5rem; max-width: 420px; }
.pipe-steps { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; }

.pipe-step {
  display: grid;
  grid-template-columns: 2.6rem 1fr;
  gap: .5rem;
  padding: 1.3rem 0;
  border-bottom: 1px solid var(--border);
}
.pipe-step:first-child { padding-top: 0; }
.pipe-step:last-child  { border-bottom: none; }
.pipe-num  { font-family: var(--mono); font-size: .64rem; color: var(--ink-3); padding-top: .1rem; }
.pipe-name { font-size: .9rem; font-weight: 600; letter-spacing: -.01em; color: var(--salmon); margin: 0 0 .28rem; }
.pipe-desc { font-size: .82rem; line-height: 1.65; color: var(--ink-2); margin: 0; }

.pipe-right { position: sticky; top: calc(var(--nav-h) + 1.5rem); }
.code-block {
  background: var(--ink);
  border-radius: 12px;
  padding: 1.55rem;
  font-family: var(--mono);
  font-size: .75rem;
  line-height: 1.95;
}
.c-lbl { display: block; font-size: .58rem; letter-spacing: .12em; text-transform: uppercase; color: rgba(242,239,233,.2); margin-bottom: .6rem; }
.c-dim { color: rgba(242,239,233,.24); }
.c-arr { color: var(--teal); }
.c-hl  { color: var(--salmon); }
.c-ok  { color: var(--teal); }
.c-txt { color: rgba(242,239,233,.7); }

/* ── Dark band (shared by privacy + install) ─────────────────────────── */
.dark-band,
.install-band {
  background: var(--ink);
  padding: clamp(4rem, 8vw, 7rem) var(--pad);
}
.dark-inner,
.install-inner {
  max-width: var(--max);
  margin: 0 auto;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 5rem;
  align-items: start;
}
.dark-label { color: rgba(242, 239, 233, .26) !important; }
.dark-h2    { color: #f2efe9 !important; }
.dark-lead  { font-size: .95rem; line-height: 1.75; color: rgba(242, 239, 233, .44); }

/* Privacy list */
.priv-list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: .6rem; }
.priv-item {
  display: flex;
  align-items: flex-start;
  gap: .85rem;
  padding: .95rem 1.1rem;
  background: rgba(255,255,255,.04);
  border: 1px solid rgba(255,255,255,.07);
  border-radius: 8px;
}
.priv-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--teal);
  flex-shrink: 0;
  margin-top: .44rem;
}
.priv-dot--accent { background: var(--salmon); }
.priv-title { font-size: .87rem; font-weight: 500; color: #f2efe9; margin: 0 0 .16rem; }
.priv-desc  { font-size: .77rem; color: rgba(242,239,233,.36); line-height: 1.6; margin: 0; }

/* ── Story ──────────────────────────────────────────────────────────── */
.story-wrap {
  display: grid;
  grid-template-columns: 1fr 1.4fr;
  gap: 5rem;
  align-items: start;
}
.story-h2 { max-width: 260px; }
.story-right p {
  font-size: .97rem;
  line-height: 1.85;
  color: var(--ink-2);
  margin: 0 0 1.35rem;
}
.story-right p:last-child { margin-bottom: 0; }
.story-right em { font-style: normal; color: var(--ink); font-weight: 500; }

/* ── Install band ───────────────────────────────────────────────────── */
.install-manifest {
  display: flex;
  flex-direction: column;
  gap: 0;
}
.manifest-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 1rem;
  padding: .6rem 0;
  border-bottom: 1px solid rgba(255,255,255,.07);
}
.manifest-row:last-child { border-bottom: none; }
.manifest-key {
  font-family: var(--mono);
  font-size: .7rem;
  color: rgba(242,239,233,.3);
  letter-spacing: .06em;
  flex-shrink: 0;
}
.manifest-val {
  font-size: .82rem;
  color: rgba(242,239,233,.65);
  text-align: right;
}

/* Install steps */
.install-steps { list-style: none; margin: 0 0 1.75rem; padding: 0; display: flex; flex-direction: column; }
.install-step {
  display: grid;
  grid-template-columns: 2.2rem 1fr;
  gap: .5rem;
  align-items: start;
  padding: 1rem 0;
  border-bottom: 1px solid rgba(255,255,255,.06);
}
.install-step:first-child { padding-top: 0; }
.install-step:last-child  { border-bottom: none; }

.is-num {
  font-family: var(--mono);
  font-size: .6rem;
  color: rgba(242,239,233,.2);
  padding-top: .85rem;
  letter-spacing: .06em;
}
.is-body    { display: flex; flex-direction: column; gap: .4rem; }
.is-comment {
  font-family: var(--mono);
  font-size: .65rem;
  color: rgba(242,239,233,.3);
  letter-spacing: .04em;
}

.is-cmd {
  display: flex;
  align-items: center;
  gap: .6rem;
  width: 100%;
  padding: .65rem .9rem;
  background: rgba(255,255,255,.05);
  border: 1px solid rgba(255,255,255,.08);
  border-radius: 7px;
  cursor: pointer;
  text-align: left;
  transition: background .18s, border-color .18s;
}
.is-cmd:hover { background: rgba(255,255,255,.09); border-color: rgba(255,255,255,.14); }
.is-cmd--copied { border-color: rgba(90,168,154,.4); background: rgba(90,168,154,.07); }

.is-prompt {
  font-family: var(--mono);
  font-size: .75rem;
  color: var(--teal);
  flex-shrink: 0;
  user-select: none;
}
.is-text {
  font-family: var(--mono);
  font-size: .75rem;
  color: rgba(242,239,233,.72);
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.is-hint {
  font-family: var(--mono);
  font-size: .62rem;
  color: rgba(242,239,233,.22);
  flex-shrink: 0;
  white-space: nowrap;
  transition: color .18s;
}
.is-cmd:hover       .is-hint { color: rgba(242,239,233,.45); }
.is-cmd--copied     .is-hint { color: var(--teal); }

.install-actions { display: flex; gap: .65rem; flex-wrap: wrap; }
.btn-install-primary {
  background: var(--bg);
  color: var(--ink);
  border: 1px solid transparent;
  transition: background .18s, color .18s;
}
.btn-install-primary:hover { background: var(--salmon); color: var(--bg); }
.btn-install-ghost {
  background: transparent;
  color: rgba(242,239,233,.55);
  border: 1px solid rgba(255,255,255,.1);
  transition: color .18s, border-color .18s;
}
.btn-install-ghost:hover { color: rgba(242,239,233,.9); border-color: rgba(255,255,255,.25); }

/* ── Footer ─────────────────────────────────────────────────────────── */
.site-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: .85rem;
  padding: 1.65rem var(--pad);
  max-width: var(--max);
  margin: 0 auto;
  border-top: 1px solid var(--border);
}
.footer-copy { font-size: .77rem; color: var(--ink-3); }
.footer-copy a { color: inherit; text-decoration: none; transition: color .15s; }
.footer-copy a:hover { color: var(--ink); }
.footer-links { display: flex; gap: 1.35rem; }
.footer-links a { font-size: .77rem; color: var(--ink-3); text-decoration: none; transition: color .15s; }
.footer-links a:hover { color: var(--salmon); }

/* ── Scroll reveal ──────────────────────────────────────────────────── */
.reveal { opacity: 0; transition: opacity .5s ease; }
.reveal.in { opacity: 1; }

/* ── Inline code inside v-html ──────────────────────────────────────── */
:deep(.ic) {
  font-family: var(--mono);
  font-size: .82em;
  color: var(--salmon);
  background: rgba(224, 107, 80, .08);
  border: 1px solid rgba(224, 107, 80, .15);
  border-radius: 4px;
  padding: .1em .35em;
}

/* ── Responsive — tablet (≤ 900px) ─────────────────────────────────── */
@media (max-width: 900px) {
  .section-header { grid-template-columns: 1fr; gap: .85rem; }
  .section-header .section-lead { max-width: 100%; }

  .cmd-grid { grid-template-columns: repeat(2, 1fr); }
  .cmd-card:nth-child(3n)        { border-right: 1px solid var(--border); }
  .cmd-card:nth-child(2n)        { border-right: none; }
  .cmd-card:nth-last-child(-n+3) { border-bottom: 1px solid var(--border); }
  .cmd-card:nth-last-child(-n+2) { border-bottom: none; }

  .pipe-wrap   { grid-template-columns: 1fr; gap: 2.5rem; }
  .pipe-right  { position: static; }
  .pipe-lead   { max-width: 100%; }

  .dark-inner,
  .install-inner { grid-template-columns: 1fr; gap: 2.5rem; }

  .story-wrap { grid-template-columns: 1fr; gap: 2rem; }
  .story-h2   { max-width: 100%; }
}

/* ── Responsive — mobile (≤ 600px) ─────────────────────────────────── */
@media (max-width: 600px) {
  .cmd-grid { grid-template-columns: 1fr; }
  .cmd-card                      { border-right: none !important; }
  .cmd-card:not(:last-child)     { border-bottom: 1px solid var(--border) !important; }
  .cmd-card:last-child           { border-bottom: none !important; }

  .hero-btns      { flex-direction: column; align-items: stretch; }
  .hero-btns .btn { justify-content: center; }

  .install-actions           { flex-direction: column; align-items: stretch; }
  .install-actions .btn      { justify-content: center; }

  .site-footer { flex-direction: column; align-items: flex-start; }
}

/* ── Responsive — small (≤ 400px) ──────────────────────────────────── */
@media (max-width: 400px) {
  .hero-tags span { font-size: .62rem; }
}

/* ── Line breaks — collapse on very small screens ───────────────────── */
.br-md { display: block; }
@media (max-width: 480px) { .br-md { display: none; } }

/* ── Reduced motion ─────────────────────────────────────────────────── */
@media (prefers-reduced-motion: reduce) {
  .reveal { opacity: 1; transition: none; }
}
</style>