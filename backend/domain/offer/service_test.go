package offer

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	appdb "github.com/RK-Consulting/skill-sifter/db"
	_ "github.com/lib/pq"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper(); _,file,_,_:=runtime.Caller(0); root:=filepath.Clean(filepath.Join(filepath.Dir(file),"../.."))
	old,_:=os.Getwd(); if err:=os.Chdir(root);err!=nil{t.Fatal(err)}; defer os.Chdir(old)
	dsn:=fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",env("TEST_DB_HOST","localhost"),env("TEST_DB_PORT","5432"),env("TEST_DB_USER","postgres"),env("TEST_DB_PASSWORD","postgres"),env("TEST_DB_NAME","skillsifter_test"))
	d,err:=sql.Open("postgres",dsn);if err!=nil{t.Skip(err)};if err:=d.Ping();err!=nil{d.Close();t.Skip(err)}
	appdb.DB=d;if err:=appdb.InitializeSchema();err!=nil{d.Close();t.Fatal(err)};return d
}
func env(k,f string)string{if v:=os.Getenv(k);v!=""{return v};return f}
func fixture(t *testing.T,d *sql.DB)(string,int,int,func()){
	t.Helper();tenant:=fmt.Sprintf("offer_test_%d",os.Getpid());must:=func(err error){if err!=nil{t.Fatal(err)}};exec:=func(q string,a ...interface{}){_,err:=d.Exec(q,a...);must(err)}
	exec("INSERT INTO companies(id,name) VALUES($1,$2)",tenant,tenant)
	var cid,client,rid int
	must(d.QueryRow("INSERT INTO candidates(name,email,tenant_id,company_name) VALUES($1,$2,$3,$4) RETURNING id","Candidate",tenant+"@c",tenant,tenant).Scan(&cid))
	must(d.QueryRow("INSERT INTO clients(name,status,tenant_id) VALUES($1,$2,$3) RETURNING id","Client","active",tenant).Scan(&client))
	must(d.QueryRow("INSERT INTO requirements(client_id,title,status,tenant_id) VALUES($1,$2,$3,$4) RETURNING id",client,"Requirement","open",tenant).Scan(&rid))
	exec("INSERT INTO recruitment_selections(tenant_id,candidate_id,requirement_id,decision) VALUES($1,$2,$3,'selected')",tenant,cid,rid)
	clean:=func(){d.Exec("DELETE FROM recruitment_offers WHERE tenant_id=$1",tenant);d.Exec("DELETE FROM recruitment_selections WHERE tenant_id=$1",tenant);d.Exec("DELETE FROM requirements WHERE tenant_id=$1",tenant);d.Exec("DELETE FROM clients WHERE tenant_id=$1",tenant);d.Exec("DELETE FROM candidates WHERE tenant_id=$1",tenant);d.Exec("DELETE FROM companies WHERE id=$1",tenant)}
	return tenant,cid,rid,clean
}
func TestService_CreateOffer(t *testing.T){d:=testDB(t);defer d.Close();tenant,cid,rid,clean:=fixture(t,d);defer clean();got,err:=NewService(NewPostgresRepository(d),d).Create(tenant,CreateInput{CandidateID:cid,RequirementID:rid});if err!=nil{t.Fatal(err)};if got.ID==0||got.Accepted{t.Fatalf("unexpected offer: %+v",got)}}
func TestService_CreateOfferRejectedSelection(t *testing.T){d:=testDB(t);defer d.Close();tenant,cid,rid,clean:=fixture(t,d);defer clean();if _,err:=d.Exec("UPDATE recruitment_selections SET decision='rejected' WHERE tenant_id=$1 AND candidate_id=$2 AND requirement_id=$3",tenant,cid,rid);err!=nil{t.Fatal(err)};_,err:=NewService(NewPostgresRepository(d),d).Create(tenant,CreateInput{CandidateID:cid,RequirementID:rid});if err!=ErrSelectionNotFound{t.Fatalf("got %v, want ErrSelectionNotFound",err)}}
func TestService_AcceptOffer(t *testing.T){d:=testDB(t);defer d.Close();tenant,cid,rid,clean:=fixture(t,d);defer clean();s:=NewService(NewPostgresRepository(d),d);if _,err:=s.Create(tenant,CreateInput{CandidateID:cid,RequirementID:rid});err!=nil{t.Fatal(err)};got,err:=s.Update(tenant,cid,rid,UpdateInput{Accepted:true});if err!=nil{t.Fatal(err)};if !got.Accepted{t.Fatalf("expected accepted offer: %+v",got)}}
func TestService_TenantIsolation(t *testing.T){d:=testDB(t);defer d.Close();tenant,cid,rid,clean:=fixture(t,d);defer clean();other:=tenant+"_other";if _,err:=d.Exec("INSERT INTO companies(id,name) VALUES($1,$2)",other,other);err!=nil{t.Fatal(err)};defer d.Exec("DELETE FROM companies WHERE id=$1",other);if _,err:=NewService(NewPostgresRepository(d),d).Get(other,cid,rid);err!=ErrNotFound{t.Fatalf("got %v, want ErrNotFound",err)}}
