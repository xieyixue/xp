# coding=utf-8
from fabric.api import env, run, local, put
env.key = '/app/keys/mac_id_rsa'
env.user = 'python_alan'
env.hosts = [
    "35.185.165.210"
]


def deploy():
    local("GOOS=linux GOARCH=amd64 go build server.go")
    put('server', '/app/first/main')
    run('sudo supervisorctl restart proxy')


def package():
    local('cd ../; tar zcvf first.tar first')


if __name__ == "__main__":
    # package()
    deploy()
