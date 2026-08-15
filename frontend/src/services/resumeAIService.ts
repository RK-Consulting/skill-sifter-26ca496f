import api from './api';

export interface ResumeResult { id:number; name:string; email:string; phone:string; skills:string; resumeFile:string; resumeStatus:string; }
export interface ResumeRecord { id:number; fileName:string; status:string; error?:string; uploadedAt:string; parsedAt?:string; name:string; email:string; phone:string; skills:string; }
export interface UploadResult { fileName:string; status:string; resumeId?:number; candidate?:{id:number;name:string;email:string;phone:string;skills:string}; error?:string; }
export interface PeriodRow { period:string; activities:number; candidates:number; resumes:number; resumeSearches:number; jobs:number; interviews:number; hires:number; businessDev:number; }
export interface ActivityRow { id:number; action:string; entityType:string; entityId?:string; description:string; actorUserId?:number; createdAt:string; }

export const resumeAIService={
  upload: async(files:File[])=>{const form=new FormData();files.forEach(f=>form.append('files',f,f.webkitRelativePath||f.name));return api.post('/resume-ai/upload',form,{headers:{'Content-Type':'multipart/form-data'}})},
  search: async(q:string)=>api.get('/resume-ai/search',{params:{q}}),
  list: async()=>api.get('/resume-ai/resumes'),
  health: async()=>api.get('/resume-ai/health'),
  periodic: async(period:string)=>api.get('/reports/periodic',{params:{period}}),
  activity: async()=>api.get('/reports/activity-log',{params:{limit:200}}),
};
