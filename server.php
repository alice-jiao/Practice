<?php
//确保在连接客户端时不会超时
set_time_limit(0);
date_default_timezone_set( 'Asia/Chongqing');  
define('WORK_PATH', dirname(__FILE__));

const IP = '127.0.0.1';
const PORT = 1993;

/*
 +-------------------------------
 *    @socket通信整个过程
 +-------------------------------
 *    @socket_create
 *    @socket_bind
 *    @socket_listen
 *    @socket_accept
 *    @socket_read
 *    @socket_write
 *    @socket_close
 +--------------------------------
 */
class Server{

    private $socket;
    private $logFilePath;


    public function __construct(){

        $this->setLogFilePath(WORK_PATH . '/logs/server.log');
        $dirName = dirname($this->logFilePath);
        if (!is_dir($dirName)) {
            $this->addLog("creating log file dir : $dirName ...");
            if(!mkdir($dirName, 0777, true)) {
                $this->addLog("Unable to create log file dir : $dirName , exiting...");
            }
        }

        if(($this->socket = socket_create(AF_INET,SOCK_STREAM,SOL_TCP)) < 0) {
            $this->addLog("socket_create() failed! reason :".socket_strerror($this->socket));
        }

        if(($bindRet = socket_bind($this->socket, IP, PORT)) < 0) {
            $this->addLog("socket_bind() failed! reason :".socket_strerror($bindRet));
        }

        if(($lisenRet = socket_listen($this->socket,4)) < 0) {
            $this->addLog("socket_listen() failed! reason :".socket_strerror($lisenRet));
        }
    }
    
    public function run(){
        $message_queue_key = ftok(__FILE__, 'a');

         $message_queue = msg_get_queue($message_queue_key, 0666);

        do {
            if (($msgsock = socket_accept($this->socket)) < 0) {
                
                $this->addLog("socket_accept() failed! reason: " . socket_strerror($msgsock));
                break;
            
            } else {

                $zhanArr = array();
                $this->addLog("parent start, pid:".getmypid());

                for ($i=0; $i <3 ; $i++) { 
                    $pid = pcntl_fork();
                    if ($pid == -1){  

                      die ("cannot fork" );  
                    
                    } else if ($pid > 0){  

                        $this->addLog("parent continue, pid:".getmypid());  

                        while (true) {

                            $buf = socket_read($msgsock,10000);
                            if(!empty($buf)){
                                $sendRet = msg_send($message_queue, 1, $buf);
                                if($sendRet){
                                    $this->addLog("send msg to queue ok!send message : ".$buf);
                                }else{
                                    $this->addLog("send msg to queue failed!!!");
                                }
                            }else{
                                $this->addLog("no message");
                                sleep(2);
                            }
                            
                        }
                        
                    } else if ($pid == 0){
                        while (true) {

                            msg_receive($message_queue, 0, $message_type, 1024, $message, true, MSG_IPC_NOWAIT);

                            if(!empty($message)){

                                $this->addLog("receive message:" . $message);
                                $this->writeFile(WORK_PATH.'/abc.txt', $message);
                                
                            }else{
                                sleep(10);
                            }  
                        }   
                        
                    }  
                }

            }

        } while (true);

    }

    /**
     * 增加一行日志,自动添加换行符
     * @param str $log 日志内容
     * @return
     */
    protected function addLog($log) {
        $log = date("Y-m-d H:i:s") . " {$log} \n";
        file_put_contents($this->logFilePath, $log, FILE_APPEND | LOCK_EX);
        echo date("Y-m-d H:i:s") . " {$log} \n";

    }

    /**
     * set log file path of task
     * @param stirng $path
     */
    protected function setLogFilePath($path) {
        $this->logFilePath = $path;
    }

    protected function writeFile($filename, $content){
        $file = fopen("$filename","a");
        while(1) {
            if (flock($file, LOCK_EX)){
                fwrite($file, $content);
                flock($file, LOCK_UN);
                fclose($file);
                break;
            } else {
               usleep(1000);
            }
        }

    }

}

$obj = new Server();
$obj->run();

?>