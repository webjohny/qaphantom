var proxyChain = require('proxy-chain');
var process = require('process');
var url = require("url");
const http = require("http");
const port = "3000";

let createTunnel = async (oldProxyUrl) => {
    if(!oldProxyUrl) {
        oldProxyUrl = 'http://1SCq278p:nkAXhVlq@45.89.19.37:12410';
    }
    let proxy = null;
    try{
        proxy = await proxyChain.anonymizeProxy(oldProxyUrl);
    }catch(e){
        console.error(e);
    }
    return proxy
};

http.createServer(async function(request, rs){
    let req = url.parse(request.url, true);
    let headers = request.headers;
    let config = require("./config");
    rs.setHeader("Content-Type", "text/html; charset=utf-8;");
    try {
        //@toDo убрать заглушку
        if (headers["X-Api-Key"] === config.key) {
            rs.write("forbidden");
        } else {
            if (req.pathname === "/create") {
                let proxy = req.query.host + ":" + req.query.port;
                if (req.query.login && req.query.pass) {
                    proxy = req.query.login + ":" + req.query.pass + "@" + proxy;
                }
                if (proxy !== "undefined:undefined") {
                    proxy = `http://${proxy}`;
                    rs.write(await createTunnel(proxy));
                } else {
                    rs.write("invalid proxy");
                }
            } else if (req.pathname === "/cancel") {
                let proxy = req.query.url;
                await proxyChain.closeAnonymizedProxy(proxy, true);
                rs.write("yes");
            } else {
                rs.statusCode = 404; // адрес не найден
                rs.write("no");
            }
        }
    }catch(e){
        console.error(e);
    }
    rs.end();
}).listen(port);
console.log(`Api server at port ${port} is running..`)