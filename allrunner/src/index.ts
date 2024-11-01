import 'dotenv/config'
import express, { json } from 'express';

import createRouter from './routes';
import AppConfig from './config';

const conf = new AppConfig()
const logger = conf.makeLogger()

const app = express();

app.use(json())
app.use(logger)

const router = createRouter(conf)
app.use(router)

app.listen(conf.port, () => {
	console.log(`app is liistening on port ${conf.port}`)
})
