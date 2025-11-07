import ws from 'k6/ws';
import { check, sleep } from 'k6';

export const options = { vus: 200, duration: '30s' };

export default function () {
  const room = 'room1';
  const url = `ws://localhost:8081/ws?room=${room}&client_id=vu${__VU}`;
  const res = ws.connect(url, {}, function (socket) {
    socket.on('open', function () { socket.send('hello'); });
    socket.on('message', function (_data) {});
    socket.setTimeout(function () { socket.close(); }, 1000 + Math.random()*2000);
  });
  check(res, { 'status is 101': (r) => r && r.status === 101 });
  sleep(0.1);
}
