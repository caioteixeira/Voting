import http from 'k6/http';
import { sleep } from 'k6';

export default function () {
    let target = ""
    if (Math.random() < 0.3) {
        target = "Carlos"
    }
    else {
        target = "Ana"
    }

    const payload = JSON.stringify({
        target: target
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    http.post('http://0.0.0.0:8080/vote', payload, params);

    sleep(0.01);
}
