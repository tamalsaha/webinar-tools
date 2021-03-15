# webinar-tools

curl -XGET http://localhost:4000/

curl -XPOST http://localhost:4000/register

# pass request body as application/x-www-form-urlencoded

curl -X POST \
  -d "first_name=Tamal&last_name=Saha&phone=+1-1234567890&job_title=CEO&work_email=tamal@appscode.com&knows_kubernetes=true" \
  http://localhost:4000/register

curl -X POST \
  -d "first_name=Tamal&last_name=Saha&phone=+1-1234567890&job_title=CEO&work_email=tamal@appscode.com&cluster_provider=aws&cluster_provider=azure&experience_level=new&marketing_reach=twitter" \
  http://localhost:4000/register

# pass request body as application/json

curl -X POST -H "Content-Type: application/json" \
  -d '{"name":"***","email":"***","product":"kubedb-community","cluster":"***","tos":"true","token":"***"}' \
  http://localhost:4000/register


{
	"first_name": "Tamal",
	"last_name": "Saha",
	"phone": "+1-1234567890",
	"job_title": "CEO",
	"work_email": "tamal@appscode.com",
	"knows_kubernetes": true
}
