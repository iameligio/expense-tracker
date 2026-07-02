import { Link } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'

export default function Landing() {
  const { user } = useAuth()

  return (
    <div className="landing">
      {/* Top bar */}
      <header className="lp-nav">
        <div className="brand"><span className="brand-mark">₱</span> Expense Tracker</div>
        <nav className="lp-nav-links">
          <a href="#features">Features</a>
          <a href="#how">How it works</a>
          {user ? (
            <Link className="btn btn-primary btn-sm" to="/dashboard">Go to dashboard</Link>
          ) : (
            <>
              <Link className="btn btn-ghost btn-sm" to="/login">Log in</Link>
              <Link className="btn btn-primary btn-sm" to="/register">Get started</Link>
            </>
          )}
        </nav>
      </header>

      {/* Hero */}
      <section className="lp-hero">
        <div className="lp-hero-copy">
          <span className="lp-badge">✨ Free to use — for a limited time</span>
          <h1>Know exactly where every <span className="accent">₱</span> goes.</h1>
          <p className="lp-sub">
            A simple monthly expense tracker built for Philippine budgets. Set your income,
            log what you spend, and watch a live dashboard tell you if you're on track to save.
          </p>
          <div className="lp-cta">
            {user ? (
              <Link className="btn btn-primary btn-lg" to="/dashboard">Open your dashboard</Link>
            ) : (
              <>
                <Link className="btn btn-primary btn-lg" to="/register">Start tracking — it's free</Link>
                <Link className="btn btn-ghost btn-lg" to="/login">I already have an account</Link>
              </>
            )}
          </div>
          <p className="lp-fineprint">No credit card. No cost. Just clarity on your money.</p>
        </div>

        {/* Product preview mock */}
        <div className="lp-preview" aria-hidden="true">
          <div className="lp-preview-card">
            <div className="lp-preview-head">
              <span className="lp-dot" /><span className="lp-dot" /><span className="lp-dot" />
              <span className="lp-preview-title">July 2026</span>
            </div>
            <div className="lp-kpis">
              <div className="lp-kpi"><span>Income</span><strong>₱50,000</strong></div>
              <div className="lp-kpi"><span>Spent</span><strong>₱26,000</strong></div>
              <div className="lp-kpi lp-kpi-good"><span>Savings</span><strong>₱24,000</strong></div>
            </div>
            <div className="lp-preview-chart">
              <div className="lp-pie" />
              <ul className="lp-legend">
                <li><i style={{ background: 'var(--fixed)' }} /> Rent</li>
                <li><i style={{ background: 'var(--variable)' }} /> Groceries</li>
                <li><i style={{ background: 'var(--wants)' }} /> Dining</li>
                <li><i style={{ background: 'var(--debts)' }} /> Loans</li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="lp-section" id="features">
        <h2 className="lp-section-title">Everything you need to budget with confidence</h2>
        <div className="lp-features">
          <Feature icon="🗂️" title="Four smart buckets"
            body="Organise spending into Fixed, Variable, Wants, and Debts — so you see the shape of your money, not just a long list." />
          <Feature icon="📊" title="A dashboard that talks back"
            body="Income, expenses, savings, and remaining budget — updated live for any month you pick." />
          <Feature icon="🥧" title="See where it really goes"
            body="A clear pie chart breaks down your spending by category or bucket at a glance." />
          <Feature icon="🎯" title="Hit your savings goal"
            body="Set a target and instantly know whether you're on track or overspending this month." />
          <Feature icon="🔒" title="Private &amp; secure"
            body="Passwords are encrypted and your data is yours alone — never shared, never sold." />
          <Feature icon="⚡" title="Fast &amp; effortless"
            body="Add an expense in seconds. No spreadsheets, no formulas, no fuss." />
        </div>
      </section>

      {/* How it works */}
      <section className="lp-section lp-how" id="how">
        <h2 className="lp-section-title">Up and running in under a minute</h2>
        <div className="lp-steps">
          <Step n="1" title="Create your free account" body="Sign up with an email and password. That's it." />
          <Step n="2" title="Set income &amp; log expenses" body="Enter your monthly income and start adding what you spend." />
          <Step n="3" title="Watch your dashboard" body="See your savings, remaining budget, and spending breakdown come alive." />
        </div>
      </section>

      {/* Final CTA */}
      <section className="lp-final">
        <h2>Ready to take control of your money?</h2>
        <p>Join now while it's completely free.</p>
        {user ? (
          <Link className="btn btn-primary btn-lg" to="/dashboard">Open your dashboard</Link>
        ) : (
          <Link className="btn btn-primary btn-lg" to="/register">Create your free account</Link>
        )}
      </section>

      <footer className="lp-footer">
        <span><span className="brand-mark">₱</span> Expense Tracker</span>
        <span className="muted">Built for Philippine budgets · Free to use for the moment</span>
      </footer>
    </div>
  )
}

function Feature({ icon, title, body }: { icon: string; title: string; body: string }) {
  return (
    <div className="lp-feature">
      <span className="lp-feature-icon">{icon}</span>
      <h3>{title}</h3>
      <p>{body}</p>
    </div>
  )
}

function Step({ n, title, body }: { n: string; title: string; body: string }) {
  return (
    <div className="lp-step">
      <span className="lp-step-n">{n}</span>
      <div>
        <h3>{title}</h3>
        <p>{body}</p>
      </div>
    </div>
  )
}
