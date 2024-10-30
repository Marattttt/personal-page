import express, { json } from 'express';
import createRouter from './routes';
import { pinoHttp } from 'pino-http';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(json())
app.use(pinoHttp())

const jsrundir = '../runtimedir'

const router = createRouter({ jsRunDir: jsrundir })
app.use(router)

app.listen(PORT, () => {
	console.log(`app is liistening on port ${PORT}`)
})
