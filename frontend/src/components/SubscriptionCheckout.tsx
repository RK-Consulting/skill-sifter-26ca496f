import { useEffect, useMemo, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui-custom/Card";
import Button from "@/components/ui-custom/Button";
import { Input } from "@/components/ui/input";
import { subscriptionService } from "@/services/api";
import { toast } from "sonner";

type Plan = { code:string; name:string; amountMinor:number; billingPeriod:string; userLimit:number };
const errorMessage = (error: unknown, fallback: string) => {
  const response = (error as { response?: { data?: { message?: string } } })?.response;
  return response?.data?.message || fallback;
};


const SubscriptionCheckout = () => {
  const [plans,setPlans]=useState<Plan[]>([]);
  const [billing,setBilling]=useState<"monthly"|"annual">("monthly");
  const [phone,setPhone]=useState("");
  const [code,setCode]=useState("");
  const [phoneSent,setPhoneSent]=useState(false);
  const [verified,setVerified]=useState(false);
  const [busy,setBusy]=useState(false);

  useEffect(()=>{ subscriptionService.getPlans().then(r=>setPlans(r.data?.data||[])).catch(()=>toast.error("Could not load plans.")); },[]);

  const visible=useMemo(()=>plans.filter(p=>p.billingPeriod===billing),[plans,billing]);

  const sendCode=async()=>{
    if(!phone.trim()){toast.error("Enter the administrator phone number.");return;}
    try{setBusy(true);await subscriptionService.sendPhoneVerification(phone.trim());setPhoneSent(true);toast.success("Phone verification code sent.");}
    catch(e){toast.error(errorMessage(e, "Could not send verification code."));}
    finally{setBusy(false);}
  };

  const verify=async()=>{
    if(code.length!==6){toast.error("Enter the 6-digit verification code.");return;}
    try{setBusy(true);await subscriptionService.verifyPhoneVerification(code);setVerified(true);toast.success("Phone number verified.");}
    catch(e){toast.error(errorMessage(e, "Verification failed."));}
    finally{setBusy(false);}
  };

  const checkout=async(planCode:string)=>{
    try{setBusy(true);const r=await subscriptionService.checkout(planCode);const url=r.data?.data?.checkoutUrl;if(!url)throw new Error("Checkout URL was not returned.");window.location.href=url;}
    catch(e){toast.error(errorMessage(e, "Could not start payment checkout."));setBusy(false);}
  };

  return <Card className="mt-6">
    <CardHeader><CardTitle>Choose a plan</CardTitle></CardHeader>
    <CardContent>
      <p className="text-sm text-muted-foreground mb-5">Your 2-day trial is free. Verify the administrator phone number before starting a paid subscription.</p>
      <div className="flex gap-2 mb-5"><Button variant={billing==="monthly"?"primary":"outline"} onClick={()=>setBilling("monthly")}>Monthly</Button><Button variant={billing==="annual"?"primary":"outline"} onClick={()=>setBilling("annual")}>Annual</Button></div>
      {!verified && <div className="rounded-xl border p-5 mb-6">
        <p className="font-medium">Phone verification</p>
        <div className="mt-3 flex flex-col gap-3 sm:flex-row"><Input placeholder="+91 98765 43210" value={phone} onChange={e=>setPhone(e.target.value)} disabled={phoneSent}/><Button onClick={sendCode} disabled={busy||phoneSent}>{phoneSent?"Code sent":"Send code"}</Button></div>
        {phoneSent && <div className="mt-3 flex flex-col gap-3 sm:flex-row"><Input inputMode="numeric" maxLength={6} placeholder="6-digit code" value={code} onChange={e=>setCode(e.target.value.replace(/\D/g,"").slice(0,6))}/><Button onClick={verify} disabled={busy}>Verify phone</Button></div>}
      </div>}
      <div className="grid gap-4 md:grid-cols-2">{visible.map(p=><div key={p.code} className="rounded-xl border p-5"><div className="font-semibold text-lg">{p.name}</div><div className="mt-2 text-2xl font-bold">₹{(p.amountMinor/100).toLocaleString("en-IN")}<span className="text-sm font-normal text-muted-foreground">/{billing==="monthly"?"month":"year"}</span></div><p className="mt-2 text-sm text-muted-foreground">Up to {p.userLimit} users</p><Button className="w-full mt-5" onClick={()=>checkout(p.code)} disabled={!verified||busy}>{verified?"Continue to payment":"Verify phone first"}</Button></div>)}</div>
    </CardContent>
  </Card>;
};

export default SubscriptionCheckout;
