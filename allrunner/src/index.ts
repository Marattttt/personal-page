import express, { json } from 'express';
import createRouter from './routes';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(json())

const jsrundir = './runtimedir'

const router = createRouter({ jsRunDir: jsrundir })
app.use(router)

app.listen(PORT, () => {
	console.log(`app is liistening on port ${PORT}`)
})
