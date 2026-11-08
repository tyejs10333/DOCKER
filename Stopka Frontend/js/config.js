
const CONFIG = {

  demo: false,


  apiUrl: '',

 
  botName: '',

 
  duelPollMs: 2000
};

const urlParams = new URLSearchParams(location.search);


const DEMO_MODE = CONFIG.demo || urlParams.has('demo');


const DEV_WITHOUT_MAX = urlParams.has('dev');
const DEV_USER_ID = urlParams.get('dev') || '1';
