
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

### Payload to retrieve report

```json
{
  "taskId": "515c760c-64e4-4511-b4ac-43587e35e542",
  "project": {
    "key": "demande-service"
  },
  "gitlab": {
    "projectId": "31",
    "ciToken": "glcbt-abc123token",
    "branchName": "sonarqube-project__sonar-server-and-gitlab-instance-interconnection",
    "branchUrl": "https://gitlab.cavom.lan/portail/demande-service/-/commits/tech/sonarqube-project__sonar-server-and-gitlab-instance-interconnection",
    "commitSha": "f6bbc79c5d1e110d00f332db63b0dae8dd006808"
  },
  "mergeRequest": {
    "iid": 29
  }
}
```
