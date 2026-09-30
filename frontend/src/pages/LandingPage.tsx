import {
  ArrowRight,
  BarChart3,
  CalendarDays,
  Check,
  ChevronDown,
  FileText,
  Search,
  ShieldCheck,
  Sparkles,
  Users,
  Zap,
} from "lucide-react";
import { useNavigate } from "react-router-dom";

const features = [
  {
    icon: FileText,
    title: "Requirements",
    text: "Capture every client requirement in one structured workspace.",
  },
  {
    icon: Search,
    title: "AI Screening & Matching",
    text: "Find the right candidates faster with intelligent screening against real requirements.",
  },
  {
    icon: Users,
    title: "Candidate Pipeline",
    text: "Move candidates cleanly from screening through submission, selection and joining.",
  },
  {
    icon: CalendarDays,
    title: "Interviews",
    text: "Keep interview scheduling, outcomes and candidate progress connected.",
  },
  {
    icon: BarChart3,
    title: "Recruitment Intelligence",
    text: "See what is moving, what is stuck and where your team needs attention.",
  },
  {
    icon: ShieldCheck,
    title: "Secure Company Workspace",
    text: "Keep your recruitment company's operational data isolated within its own tenant.",
  },
];

const plans = [
  {
    name: "Starter",
    monthly: "₹999",
    annual: "₹9,990",
    users: "3 users",
    requirements: "25 active requirements",
    candidates: "2,500 candidates",
  },
  {
    name: "Professional",
    monthly: "₹2,499",
    annual: "₹24,990",
    users: "10 users",
    requirements: "100 active requirements",
    candidates: "10,000 candidates",
    popular: true,
  },
];

const workflow = [
  "Requirements",
  "Candidates",
  "Screening",
  "Submission",
  "Interviews",
  "Selection",
  "Offer & Joining",
  "Billing",
];

const LandingPage = () => {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen overflow-x-hidden bg-slate-950 text-white">
      <header className="sticky top-0 z-50 border-b border-white/10 bg-slate-950/85 backdrop-blur-xl">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
          <button
            onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })}
            className="flex items-center gap-3"
            aria-label="SkillSifter home"
          >
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-500 shadow-lg shadow-blue-500/25">
              <Sparkles className="h-5 w-5" />
            </div>
            <div className="text-left">
              <div className="text-lg font-bold tracking-tight">SkillSifter</div>
              <div className="text-[11px] font-medium uppercase tracking-[0.16em] text-slate-500">
                Recruit. Smarter. Faster.
              </div>
            </div>
          </button>

          <nav className="hidden items-center gap-8 text-sm text-slate-300 md:flex">
            <a href="#features" className="transition hover:text-white">
              Product
            </a>
            <a href="#features" className="transition hover:text-white">
              Features
            </a>
            <a href="#workflow" className="transition hover:text-white">
              How it works
            </a>
            <a href="#pricing" className="transition hover:text-white">
              Pricing
            </a>
          </nav>

          <div className="flex items-center gap-2 sm:gap-3">
            <button
              onClick={() => navigate("/login")}
              className="rounded-lg px-3 py-2 text-sm font-medium text-slate-200 transition hover:bg-white/10 sm:px-4"
            >
              Login
            </button>
            <button
              onClick={() => navigate("/register")}
              className="rounded-lg bg-white px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-slate-100"
            >
              Start Free Trial
            </button>
          </div>
        </div>
      </header>

      <main>
        <section className="relative overflow-hidden">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_72%_10%,rgba(59,130,246,0.24),transparent_32%),radial-gradient(circle_at_20%_0%,rgba(99,102,241,0.18),transparent_34%)]" />
          <div className="relative mx-auto grid max-w-7xl gap-14 px-6 pb-24 pt-20 lg:grid-cols-[1.02fr_.98fr] lg:items-center lg:pb-28 lg:pt-28">
            <div>
              <div className="mb-7 inline-flex items-center gap-2 rounded-full border border-blue-400/20 bg-blue-400/10 px-4 py-2 text-sm font-medium text-blue-200">
                <Zap className="h-4 w-4" />
                Built for recruitment companies
              </div>

              <h1 className="max-w-3xl text-5xl font-bold leading-[1.02] tracking-[-0.035em] sm:text-6xl lg:text-7xl">
                Recruit smarter.
                <span className="block text-blue-400">Grow faster.</span>
              </h1>

              <p className="mt-7 max-w-2xl text-lg leading-8 text-slate-300 sm:text-xl">
                End-to-end recruitment software for recruitment companies. Run
                your workflow from Requirement to Joining and Billing in one
                simple platform.
              </p>

              <div className="mt-9 flex flex-wrap gap-4">
                <button
                  onClick={() => navigate("/register")}
                  className="inline-flex items-center gap-2 rounded-xl bg-blue-500 px-6 py-3.5 font-semibold shadow-xl shadow-blue-500/20 transition hover:bg-blue-400"
                >
                  Start Your 2-Day Free Trial
                  <ArrowRight className="h-4 w-4" />
                </button>
                <a
                  href="#workflow"
                  className="rounded-xl border border-white/15 px-6 py-3.5 font-semibold text-slate-200 transition hover:bg-white/5"
                >
                  See How It Works
                </a>
              </div>

              <div className="mt-5 flex flex-wrap gap-x-6 gap-y-2 text-sm text-slate-500">
                <span>✓ No credit card for trial</span>
                <span>✓ Full platform access</span>
                <span>✓ Setup in minutes</span>
              </div>
            </div>

            <div className="relative">
              <div className="absolute -inset-8 rounded-full bg-blue-500/10 blur-3xl" />
              <div className="relative rounded-[28px] border border-white/10 bg-white/[0.07] p-2 shadow-2xl shadow-blue-950/50">
                <div className="overflow-hidden rounded-[22px] border border-white/10 bg-slate-900">
                  <div className="flex items-center justify-between border-b border-white/10 px-5 py-4">
                    <div>
                      <p className="text-xs text-slate-500">SkillSifter</p>
                      <p className="mt-1 text-sm font-semibold">Recruitment Dashboard</p>
                    </div>
                    <div className="flex items-center gap-2 rounded-full bg-emerald-400/10 px-3 py-1 text-xs font-medium text-emerald-300">
                      <span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />
                      Live
                    </div>
                  </div>

                  <div className="p-5 sm:p-6">
                    <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
                      {[
                        ["Requirements", "24"],
                        ["Candidates", "486"],
                        ["Interviews", "18"],
                        ["Joined", "7"],
                      ].map(([label, value]) => (
                        <div
                          key={label}
                          className="rounded-xl border border-white/10 bg-white/[0.035] p-4"
                        >
                          <p className="text-[11px] text-slate-500">{label}</p>
                          <p className="mt-2 text-2xl font-bold">{value}</p>
                        </div>
                      ))}
                    </div>

                    <div className="mt-4 rounded-xl border border-white/10 bg-white/[0.035] p-4">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-sm font-medium">Recruitment pipeline</p>
                          <p className="mt-1 text-xs text-slate-500">This month</p>
                        </div>
                        <BarChart3 className="h-4 w-4 text-blue-400" />
                      </div>

                      <div className="mt-5 space-y-3">
                        {[
                          ["Requirements", "86%"],
                          ["Screening", "72%"],
                          ["Submission", "58%"],
                          ["Interview", "44%"],
                        ].map(([label, width]) => (
                          <div key={label}>
                            <div className="mb-1.5 flex justify-between text-[11px] text-slate-400">
                              <span>{label}</span>
                              <span>{width}</span>
                            </div>
                            <div className="h-2 rounded-full bg-white/10">
                              <div
                                className="h-2 rounded-full bg-blue-500"
                                style={{ width }}
                              />
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>

                    <div className="mt-4 grid grid-cols-3 gap-3">
                      {["AI Screening", "Client Submission", "Joining"].map(
                        (label) => (
                          <div
                            key={label}
                            className="rounded-xl border border-white/10 bg-white/[0.025] px-3 py-3 text-center text-[11px] text-slate-400"
                          >
                            {label}
                          </div>
                        ),
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="border-y border-white/10 bg-white/[0.025]">
          <div className="mx-auto flex max-w-6xl flex-wrap items-center justify-center gap-x-10 gap-y-3 px-6 py-7 text-sm font-medium text-slate-500 sm:gap-x-12">
            {workflow.map((item) => (
              <span key={item}>{item}</span>
            ))}
          </div>
        </section>

        <section id="features" className="mx-auto max-w-7xl px-6 py-24 lg:py-28">
          <div className="max-w-3xl">
            <p className="text-sm font-semibold uppercase tracking-[0.2em] text-blue-400">
              One recruitment platform
            </p>
            <h2 className="mt-4 text-4xl font-bold tracking-tight sm:text-5xl">
              Everything your recruitment team needs to move candidates forward.
            </h2>
            <p className="mt-5 text-lg leading-8 text-slate-400">
              SkillSifter is built around the actual recruitment lifecycle —
              without turning every action into another module, workflow or
              charge.
            </p>
          </div>

          <div className="mt-14 grid gap-5 md:grid-cols-2 lg:grid-cols-3">
            {features.map(({ icon: Icon, title, text }) => (
              <div
                key={title}
                className="rounded-2xl border border-white/10 bg-white/[0.035] p-7 transition hover:-translate-y-1 hover:bg-white/[0.06]"
              >
                <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-blue-500/10 text-blue-300">
                  <Icon className="h-5 w-5" />
                </div>
                <h3 className="mt-5 text-lg font-semibold">{title}</h3>
                <p className="mt-2 leading-7 text-slate-400">{text}</p>
              </div>
            ))}
          </div>
        </section>

        <section id="workflow" className="border-y border-white/10 bg-white/[0.025] py-24 lg:py-28">
          <div className="mx-auto max-w-7xl px-6">
            <div className="text-center">
              <p className="text-sm font-semibold uppercase tracking-[0.2em] text-blue-400">
                Simple by design
              </p>
              <h2 className="mt-4 text-4xl font-bold tracking-tight sm:text-5xl">
                One flow from requirement to billing.
              </h2>
              <p className="mx-auto mt-5 max-w-2xl text-lg leading-8 text-slate-400">
                Keep the entire recruitment lifecycle connected without
                stitching together multiple tools.
              </p>
            </div>

            <div className="mx-auto mt-14 grid max-w-6xl gap-3 sm:grid-cols-2 lg:grid-cols-4">
              {[
                ["01", "Requirement", "Capture the client need."],
                ["02", "Candidates", "Source and screen talent."],
                ["03", "Submission", "Interview and select."],
                ["04", "Joining → Billing", "Close the recruitment cycle."],
              ].map(([number, title, text], index) => (
                <div
                  key={number}
                  className="relative rounded-2xl border border-white/10 bg-slate-950 p-6"
                >
                  <div className="text-sm font-semibold text-blue-400">{number}</div>
                  <h3 className="mt-3 font-semibold">{title}</h3>
                  <p className="mt-2 text-sm leading-6 text-slate-500">{text}</p>
                  {index < 3 && (
                    <ArrowRight className="absolute -right-5 top-1/2 z-10 hidden h-5 w-5 text-slate-600 lg:block" />
                  )}
                </div>
              ))}
            </div>
          </div>
        </section>

        <section id="pricing" className="mx-auto max-w-7xl px-6 py-24 lg:py-28">
          <div className="text-center">
            <p className="text-sm font-semibold uppercase tracking-[0.2em] text-blue-400">
              Simple SaaS pricing
            </p>
            <h2 className="mt-4 text-4xl font-bold tracking-tight sm:text-5xl">
              Pay for the platform,
              <span className="block text-blue-400">not for every click.</span>
            </h2>
            <p className="mx-auto mt-5 max-w-2xl text-lg leading-8 text-slate-400">
              One predictable subscription for your recruitment operation.
              No per-candidate, per-job or per-action charges.
            </p>
          </div>

          <div className="mx-auto mt-14 grid max-w-4xl gap-6 md:grid-cols-2">
            {plans.map((plan) => (
              <div
                key={plan.name}
                className={`relative rounded-3xl border p-8 ${plan.popular ? "border-blue-400/50 bg-blue-500/[0.08]" : "border-white/10 bg-white/[0.035]"}`}
              >
                {plan.popular && (
                  <div className="absolute right-6 top-6 rounded-full bg-blue-500 px-3 py-1 text-xs font-semibold">
                    Most chosen
                  </div>
                )}

                <h3 className="text-2xl font-bold">{plan.name}</h3>
                <p className="mt-6 text-4xl font-bold">
                  {plan.monthly}
                  <span className="text-base font-normal text-slate-500">
                    {" "}
                    / month
                  </span>
                </p>
                <p className="mt-2 text-sm text-slate-500">
                  {plan.annual} / year
                </p>

                <div className="mt-8 space-y-3 text-sm text-slate-300">
                  {[
                    plan.users,
                    plan.requirements,
                    plan.candidates,
                    "Complete recruitment workflow",
                    "AI screening & matching",
                    "No per-click charges",
                  ].map((item) => (
                    <div key={item} className="flex gap-3">
                      <Check className="h-5 w-5 shrink-0 text-emerald-400" />
                      <span>{item}</span>
                    </div>
                  ))}
                </div>

                <button
                  onClick={() => navigate("/register")}
                  className={`mt-9 w-full rounded-xl px-5 py-3.5 font-semibold transition ${plan.popular ? "bg-blue-500 hover:bg-blue-400" : "bg-white/10 hover:bg-white/15"}`}
                >
                  Start 2-Day Free Trial
                </button>
              </div>
            ))}
          </div>

          <div className="mx-auto mt-10 max-w-3xl rounded-2xl border border-blue-400/15 bg-blue-400/[0.06] px-6 py-5 text-center">
            <p className="font-semibold text-blue-100">
              Pay for the platform, not for every click.
            </p>
            <p className="mt-1 text-sm leading-6 text-slate-400">
              Your team should be rewarded for recruiting more, not charged
              every time they use the software.
            </p>
          </div>

          <p className="mt-6 text-center text-sm text-slate-500">
            GST applicable where required.
          </p>
        </section>

        <section className="border-t border-white/10 bg-blue-500/[0.06] py-24">
          <div className="mx-auto max-w-4xl px-6 text-center">
            <p className="text-sm font-semibold uppercase tracking-[0.2em] text-blue-300">
              Start simple
            </p>
            <h2 className="mt-4 text-4xl font-bold tracking-tight sm:text-5xl">
              Ready to run your recruitment company on SkillSifter?
            </h2>
            <p className="mx-auto mt-5 max-w-2xl text-lg leading-8 text-slate-400">
              Create your company account, verify your email and get full
              platform access for two days.
            </p>
            <button
              onClick={() => navigate("/register")}
              className="mt-8 inline-flex items-center gap-2 rounded-xl bg-white px-7 py-3.5 font-semibold text-slate-950 transition hover:bg-slate-100"
            >
              Start Free Trial
              <ArrowRight className="h-4 w-4" />
            </button>
          </div>
        </section>
      </main>

      <footer className="border-t border-white/10">
        <div className="mx-auto flex max-w-7xl flex-col gap-5 px-6 py-10 text-sm text-slate-500 md:flex-row md:items-center md:justify-between">
          <div>
            <span className="font-semibold text-slate-300">SkillSifter</span>
            <span className="mx-2">·</span>
            Recruitment Intelligence Platform
          </div>
          <div className="flex flex-wrap gap-6">
            <a href="#features" className="transition hover:text-white">
              Features
            </a>
            <a href="#workflow" className="transition hover:text-white">
              How it works
            </a>
            <a href="#pricing" className="transition hover:text-white">
              Pricing
            </a>
            <button
              onClick={() => navigate("/login")}
              className="transition hover:text-white"
            >
              Login
            </button>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default LandingPage;
