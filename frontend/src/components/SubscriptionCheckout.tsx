import { useEffect, useMemo, useState } from "react";
import { Card, CardContent } from "@/components/ui-custom/Card";
import Button from "@/components/ui-custom/Button";
import { Input } from "@/components/ui/input";
import { subscriptionService } from "@/services/api";
import { toast } from "sonner";

type Plan = { code: string; name: string; amountMinor: number; billingPeriod: string; userLimit: number };
const errorMessage = (error: unknown, fallback: string) => {
  const response = (error as { response?: { data?: { message?: string } } })?.response;
  return response?.data?.message || fallback;
};

const SubscriptionCheckout = () => {
  const [plans, setPlans] = useState<Plan[]>([]);
  const [billing, setBilling] = useState<"monthly" | "annual">("monthly");
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [phoneSent, setPhoneSent] = useState(false);
  const [verified, setVerified] = useState(false);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    subscriptionService.getPlans()
      .then((r) => setPlans(r.data?.data || []))
      .catch(() => toast.error("Could not load plans."));
  }, []);

  const visible = useMemo(() => plans.filter((p) => p.billingPeriod === billing), [plans, billing]);

  const sendCode = async () => {
    if (!phone.trim()) {
      toast.error("Enter the administrator phone number.");
      return;
    }
    try {
      setBusy(true);
      await subscriptionService.sendPhoneVerification(phone.trim());
      setPhoneSent(true);
      toast.success("Phone verification code sent.");
    } catch (e) {
      toast.error(errorMessage(e, "Could not send verification code."));
    } finally {
      setBusy(false);
    }
  };

  const verify = async () => {
    if (code.length !== 6) {
      toast.error("Enter the 6-digit verification code.");
      return;
    }
    try {
      setBusy(true);
      await subscriptionService.verifyPhoneVerification(code);
      setVerified(true);
      toast.success("Phone number verified.");
    } catch (e) {
      toast.error(errorMessage(e, "Verification failed."));
    } finally {
      setBusy(false);
    }
  };

  const checkout = async (planCode: string) => {
    try {
      setBusy(true);
      const r = await subscriptionService.checkout(planCode);
      const url = r.data?.data?.checkoutUrl;
      if (!url) throw new Error("Checkout URL was not returned.");
      window.location.href = url;
    } catch (e) {
      toast.error(errorMessage(e, "Could not start payment checkout."));
      setBusy(false);
    }
  };

  return (
    <Card className="mt-8 overflow-hidden border-slate-200 shadow-sm">
      <div className="bg-slate-950 px-6 py-7 text-white sm:px-8">
        <div className="max-w-3xl">
          <div className="text-xs font-semibold uppercase tracking-[0.18em] text-blue-300">Subscription</div>
          <h2 className="mt-2 text-2xl font-bold tracking-tight sm:text-3xl">Choose the right plan for your team</h2>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-300">
            Your 2-day trial is free. When you are ready, verify the administrator phone number and continue securely to payment.
          </p>
          <div className="mt-4 inline-flex items-center rounded-full border border-emerald-400/20 bg-emerald-400/10 px-3 py-1.5 text-xs font-medium text-emerald-200">
            Pay for the platform, not for every click.
          </div>
        </div>
      </div>

      <CardContent className="p-6 sm:p-8">
        <div className="flex flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="font-semibold text-slate-900">Billing period</div>
            <div className="mt-1 text-sm text-slate-500">Annual plans are billed once per year.</div>
          </div>
          <div className="inline-flex rounded-xl bg-slate-100 p-1">
            {(["monthly", "annual"] as const).map((period) => (
              <button
                key={period}
                type="button"
                onClick={() => setBilling(period)}
                className={`rounded-lg px-4 py-2 text-sm font-semibold transition ${billing === period ? "bg-white text-slate-900 shadow-sm" : "text-slate-500 hover:text-slate-900"}`}
              >
                {period === "monthly" ? "Monthly" : "Annual"}
              </button>
            ))}
          </div>
        </div>

        {!verified && (
          <div className="mt-7 rounded-2xl border border-blue-100 bg-blue-50/60 p-5 sm:p-6">
            <div className="flex gap-4">
              <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-blue-600 text-sm font-bold text-white">1</div>
              <div className="w-full">
                <div className="font-semibold text-slate-900">Verify administrator phone</div>
                <p className="mt-1 text-sm text-slate-500">This protects subscription changes for your company account.</p>
                <div className="mt-4 flex flex-col gap-3 sm:flex-row">
                  <Input
                    placeholder="+91 98765 43210"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    disabled={phoneSent}
                    className="bg-white"
                  />
                  <Button onClick={sendCode} disabled={busy || phoneSent} className="sm:min-w-32">
                    {phoneSent ? "Code sent" : "Send code"}
                  </Button>
                </div>
                {phoneSent && (
                  <div className="mt-3 flex flex-col gap-3 sm:flex-row">
                    <Input
                      inputMode="numeric"
                      maxLength={6}
                      placeholder="Enter 6-digit code"
                      value={code}
                      onChange={(e) => setCode(e.target.value.replace(/D/g, "").slice(0, 6))}
                      className="bg-white tracking-[0.25em]"
                    />
                    <Button onClick={verify} disabled={busy} variant="outline" className="sm:min-w-32">
                      Verify phone
                    </Button>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {verified && (
          <div className="mt-7 flex items-center gap-3 rounded-2xl border border-emerald-100 bg-emerald-50 px-5 py-4">
            <span className="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-600 text-sm font-bold text-white">✓</span>
            <div>
              <div className="font-semibold text-emerald-900">Phone verified</div>
              <div className="text-sm text-emerald-700">You can now continue to secure payment.</div>
            </div>
          </div>
        )}

        <div className="mt-8 grid gap-5 md:grid-cols-2">
          {visible.map((p, index) => (
            <div
              key={p.code}
              className={`relative rounded-2xl border p-6 transition-all hover:-translate-y-0.5 hover:shadow-lg ${index === 1 ? "border-blue-200 ring-1 ring-blue-100" : "border-slate-200"}`}
            >
              {index === 1 && (
                <span className="absolute right-5 top-5 rounded-full bg-blue-600 px-2.5 py-1 text-[10px] font-bold uppercase tracking-wide text-white">
                  Popular
                </span>
              )}
              <div className="text-sm font-medium text-slate-500">{p.name}</div>
              <div className="mt-3 flex items-end gap-1">
                <span className="text-3xl font-bold tracking-tight text-slate-950">₹{(p.amountMinor / 100).toLocaleString("en-IN")}</span>
                <span className="pb-1 text-sm text-slate-500">/{billing === "monthly" ? "month" : "year"}</span>
              </div>
              <div className="mt-2 text-sm text-slate-500">Up to {p.userLimit} users</div>
              <div className="mt-5 space-y-2 text-sm text-slate-700">
                <div>✓ Full recruitment workflow</div>
                <div>✓ Company workspace</div>
                <div>✓ No per-click charges</div>
              </div>
              <Button className="mt-6 w-full" onClick={() => checkout(p.code)} disabled={!verified || busy}>
                {verified ? "Continue to payment" : "Verify phone first"}
              </Button>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
};

export default SubscriptionCheckout;
