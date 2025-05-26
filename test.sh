curl -X POST http://localhost:8081/run \
     -H "Content-Type: application/json" \
     -d '{"code": "你的Go代码"}'


curl -X POST http://localhost:8081/run  \
-H "Content-Type: application/json" -d '{"code": "package main\n\nfunc main() {\n    undefined_function()\n}"}'
