import Elysia from "elysia";

const task = new Elysia({ prefix: "/task" }).ws("/", {
  message(ws, message) {
    ws.send(message);
  },
});

export default task;
