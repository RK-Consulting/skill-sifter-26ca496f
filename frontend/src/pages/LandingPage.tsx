import { ArrowRight, Check, ChevronDown, Sparkles, Users, FileText, Search, CalendarDays, BarChart3, ShieldCheck, Zap } from "lucide-react";
import { useNavigate } from "react-router-dom";

const features = [
  { icon: FileText, title: "Requirements", text: "Capture every client requirement in one structured workspace." },
  { icon: Search, title: "AI Screening", text: "Find and screen candidates against real recruitment requirements." },
  { icon: Users, title: "Candidate Pipeline", text: "Move candidates from screening to submission, selection and joining." },
  { icon: CalendarDays, title: "Interviews", text: "Keep interview scheduling and outcomes connected to the recruitment flow." },
  { icon: BarChart3, title: "Recruitment Intelligence", text: "See what is moving, what is stuck and where your pipeline needs attention." },
  { icon: ShieldCheck, title: "Company Isolation", text: "Your recruitment company's data stays within its own tenant." },
];

const plans = [
  { name: "Starter", monthly: "₹999", annual: "₹9,990", users: "3 users", requirements: "25 active requirements", candidates: "2,500 candidates" },
  { name: "Professional", monthly: "₹2,499", annual: "₹24,990", users: "10 users", requirements: "100 active requirements", candidates: "10,000 candidates", popular: true },
];

const LandingPage = () => {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-slate-950 text-white">
      <header className="sticky top-0 z-50 border-b border-white/10 bg-slate-950/90 backdrop-blur">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
          <button onClick={() => window.scrollTo({ top: 0, behavior: "smooth" })} className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-blue-500 shadow-lg shadow-blue-500/20">
              <Sparkles className="h-5 w-5" />
            </div>
            <div className="text-left">
              <div className="text-lg font-bold tracking-tight">SkillSifter</div>
              <div className="text-xs text-slate-400">Recruitment Intelligence</div>
            </div>
          </button>
          <nav className="hidden items-center gap-8 text-sm text-slate-300 md:flex">
            <a href="#features" className="hover:text-white">Features</a>
            <a href="#workflow" className="hover:text-white">How it works</a>
            <a href="#pricing" className="hover:text-white">Pricing</a>
          </nav>
          <div className="flex items-center gap-3">
            <button onClick={() => navigate("/login")} className="rounded-lg px-4 py-2 text-sm font-medium text-slate-200 hover:bg-white/10">Login</button>
            <button onClick={() => navigate("/register")} className="rounded-lg bg-white px-4 py-2 text-sm font-semibold text-slate-950 hover:bg-slate-100">Start free trial</button>
          </div>
        </div>
      </header>

      <main>
        <section className="relative overflow-hidden">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_0%,rgba(59,130,246,0.22),transparent_45%)]" />
          <div className="relative mx-auto grid max-w-7xl gap-14 px-6 pb-24 pt-24 lg:grid-cols-[1.05fr_.95fr] lg:items-center lg:pt-32">
            <div>
              <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-blue-400/20 bg-blue-400/10 px-4 py-2 text-sm text-blue-200">
                <Zap className="h-4 w-4" /> Built for recruitment companies
              </div>
              <h1 className="max-w-4xl text-5xl font-bold leading-[1.05] tracking-tight sm:text-6xl">
                Recruitment software that follows the work, not the paperwork.
              </h1>
              <p className="mt-7 max-w-2xl text-lg leading-8 text-slate-300">
                Manage Requirements, Candidates, Screening, Submission, Feedback, Interviews, Selection, Offers, Joining and Billing in one simple recruitment platform.
              </p>
              <div className="mt-9 flex flex-wrap gap-4">
                <button onClick={() => navigate("/register")} className="inline-flex items-center gap-2 rounded-xl bg-blue-500 px-6 py-3.5 font-semibold shadow-xl shadow-blue-500/20 hover:bg-blue-400">
                  Start your 2-day free trial <ArrowRight className="h-4 w-4" />
                </button>
                <a href="#pricing" className="rounded-xl border border-white/15 px-6 py-3.5 font-semibold text-slate-200 hover:bg-white/5">See pricing</a>
              </div>
              <p className="mt-4 text-sm text-slate-500">No credit card for the trial · Simple company pricing · India-ready billing</p>
            </div>

            <div className="rounded-3xl border border-white/10 bg-white/[0.06] p-3 shadow-2xl shadow-blue-950/40">
              <div className="rounded-2xl border border-white/10 bg-slate-900 p-6">
                <div className="flex items-center justify-between border-b border-white/10 pb-5">
                  <div><p className="text-sm text-slate-400">Recruitment workspace</p><h2 className="mt-1 text-xl font-semibold">Today's pipeline</h2></div>
                  <div className="rounded-lg bg-emerald-400/10 px-3 py-1.5 text-xs font-medium text-emerald-300">Live</div>
                </div>
                <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
                  {[["Requirements","24"],["Candidates","486"],["Interviews","18"],["Joined","7"]].map(([label,value]) => (
                    <div key={label} className="rounded-xl border border-white/10 bg-white/[0.03] p-4">
                      <p className="text-xs text-slate-500">{label}</p><p className="mt-2 text-2xl font-bold">{value}</p>
                    </div>
                  ))}
                </div>
                <div className="mt-5 rounded-xl border border-white/10 bg-white/[0.03] p-4">
                  <div className="mb-4 flex justify-between text-sm"><span className="text-slate-300">Recruitment flow</span><span className="text-blue-300">This month</span></div>
                  <div className="flex items-center gap-2 text-xs text-slate-400">
                    {["Requirement","Screening","Submission","Interview","Selection","Joined"].map((x,i)=><div key={x} className="flex flex-1 items-center gap-2"><div className="h-2 flex-1 rounded-full bg-blue-500/70" /><span className="hidden xl:block">{i+1}</span></div>)}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="border-y border-white/10 bg-white/[0.025] py-8">
          <div className="mx-auto flex max-w-5xl flex-wrap items-center justify-center gap-x-12 gap-y-4 px-6 text-sm text-slate-400">
            <span>Requirements</span><span>AI Screening</span><span>Candidate Pipeline</span><span>Interviews</span><span>Offers</span><span>Joining</span><span>Billing</span>
          </div>
        </section>

        <section id="features" className="mx-auto max-w-7xl px-6 py-24">
          <div className="max-w-2xl"><p className="text-sm font-semibold uppercase tracking-[0.2em] text-blue-400">One recruitment platform</p><h2 className="mt-4 text-4xl font-bold tracking-tight">Everything your recruitment team needs to move candidates forward.</h2><p className="mt-5 text-lg text-slate-400">SkillSifter is designed around the actual recruitment lifecycle — without turning every action into another module or another bill.</p></div>
          <div className="mt-14 grid gap-5 md:grid-cols-2 lg:grid-cols-3">
            {features.map(({icon:Icon,title,text}) => <div key={title} className="rounded-2xl border border-white/10 bg-white/[0.035] p-7 hover:bg-white/[0.06]"><div className="flex h-11 w-11 items-center justify-center rounded-xl bg-blue-500/10 text-blue-300"><Icon className="h-5 w-5"/></div><h3 className="mt-5 text-lg font-semibold">{title}</h3><p className="mt-2 leading-7 text-slate-400">{text}</p></div>)}
          </div>
        </section>

        <section id="workflow" className="bg-white/[0.025] py-24">
          <div className="mx-auto max-w-7xl px-6">
            <div className="text-center"><p className="text-sm font-semibold uppercase tracking-[0.2em] text-blue-400">Simple by design</p><h2 className="mt-4 text-4xl font-bold">One flow from requirement to billing.</h2></div>
            <div className="mx-auto mt-14 grid max-w-5xl gap-3 md:grid-cols-4">
              {["Requirement","Candidate & Screening","Submission & Interview","Selection → Offer → Joining → Billing"].map((x,i) => <div key={x} className="relative rounded-2xl border border-white/10 bg-slate-950 p-6"><div className="text-sm font-semibold text-blue-300">0{i+1}</div><p className="mt-3 font-medium">{x}</p>{i<3 && <ArrowRight className="absolute -right-5 top-1/2 hidden h-5 w-5 text-slate-600 md:block" />}</div>)}
            </div>
          </div>
        </section>

        <section id="pricing" className="mx-auto max-w-7xl px-6 py-24">
          <div className="text-center"><p className="text-sm font-semibold uppercase tracking-[0.2em] text-blue-400">Simple SaaS pricing</p><h2 className="mt-4 text-4xl font-bold">Pay for the platform, not for every click.</h2><p className="mx-auto mt-5 max-w-2xl text-slate-400">Two plans. One predictable subscription. No per-candidate, per-job or per-action charges.</p></div>
          <div className="mx-auto mt-14 grid max-w-4xl gap-6 md:grid-cols-2">
            {plans.map(plan => <div key={plan.name} className={`relative rounded-3xl border ${plan.popular ? "border-blue-400/50 bg-blue-500/[0.08]" : "border-white/10 bg-white/[0.035]"} p-8`}>{plan.popular && <div className="absolute right-6 top-6 rounded-full bg-blue-500 px-3 py-1 text-xs font-semibold">Professional</div>}<h3 className="text-2xl font-bold">{plan.name}</h3><p className="mt-5 text-4xl font-bold">{plan.monthly}<span className="text-base font-normal text-slate-500"> / month</span></p><p className="mt-2 text-sm text-slate-500">{plan.annual} / year</p><div className="mt-7 space-y-3 text-sm text-slate-300">{[plan.users,plan.requirements,plan.candidates,"Complete recruitment workflow","AI screening","No per-click charges"].map(x=><div key={x} className="flex gap-3"><Check className="h-5 w-5 shrink-0 text-emerald-400"/>{x}</div>)}</div><button onClick={() => navigate("/register")} className={`mt-8 w-full rounded-xl px-5 py-3 font-semibold ${plan.popular ? "bg-blue-500 hover:bg-blue-400" : "bg-white/10 hover:bg-white/15"}`}>Start 2-day trial</button></div>)}
          </div>
          <p className="mt-6 text-center text-sm text-slate-500">GST applicable where required. Third plan will be introduced only when customer demand justifies it.</p>
        </section>

        <section className="border-t border-white/10 bg-blue-500/[0.06] py-24">
          <div className="mx-auto max-w-4xl px-6 text-center"><h2 className="text-4xl font-bold">Ready to run your recruitment company on SkillSifter?</h2><p className="mt-5 text-lg text-slate-400">Create your company account, verify your email and get a full 2-day trial.</p><button onClick={() => navigate("/register")} className="mt-8 inline-flex items-center gap-2 rounded-xl bg-white px-7 py-3.5 font-semibold text-slate-950 hover:bg-slate-100">Start free trial <ArrowRight className="h-4 w-4"/></button></div>
        </section>
      </main>

      <footer className="border-t border-white/10">
        <div className="mx-auto flex max-w-7xl flex-col gap-4 px-6 py-10 text-sm text-slate-500 md:flex-row md:items-center md:justify-between">
          <div><span className="font-semibold text-slate-300">SkillSifter</span> · Recruitment Intelligence Platform</div>
          <div className="flex gap-6"><a href="#features" className="hover:text-white">Features</a><a href="#pricing" className="hover:text-white">Pricing</a><button onClick={() => navigate("/login")} className="hover:text-white">Login</button></div>
        </div>
      </footer>
    </div>
  );
};

export default LandingPage;
