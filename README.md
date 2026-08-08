
### The postman pre-script for test
```js
const crypto = require("crypto-js");
const payload = JSON.stringify(JSON.parse(pm.request.body ? pm.request.body.raw : ""));
const ts = new Date().getTime();
sign = `${pm.request.method}
${pm.request.url.getPath()}
${payload}
${ts}`;

const signature = crypto.HmacSHA256(sign, "<webhook-secret>").toString(crypto.enc.Hex);

pm.request.addHeader({key: "x-sonar-webhook-hmac-sign", value: signature})
pm.request.addHeader({key: "x-sonar-webhook-Timestamp", value: ts})
```
