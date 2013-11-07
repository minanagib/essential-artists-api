<?php

include(__DIR__ . '/config.php');
use PhpAmqpLib\Connection\AMQPConnection;
use PhpAmqpLib\Message\AMQPMessage;
require_once "/Users/minanaguib/Downloads/thrift-0.9.1/lib/php/lib/Thrift/ClassLoader/ThriftClassLoader.php";
use Thrift\ClassLoader\ThriftClassLoader;
use Thrift\Protocol\TBinaryProtocol;
use Thrift\Serializer\TBinarySerializer;

$loader = new ThriftClassLoader();
$loader->registerNamespace('Thrift','/Users/minanaguib/Downloads/thrift-0.9.1//lib/php/lib');
$GEN_DIR = '/Users/minanaguib/development/go/src/crowdsurge/messages/gen-php';

$loader->registerDefinition('Order', $GEN_DIR);
$loader->registerDefinition('OrderLine', $GEN_DIR);
$loader->register();



# Create the Order Object (Thrift)
$order = new Order\Order();

# Create some OrderLine Objects (Thrift)
$order->orderLineList = array(new OrderLine\OrderLine(array(
                                       'orderId'  => '12345',
                                       'productId'=> 'SPIDERMAN')),
                              new OrderLine\OrderLine(array(
                                       'orderId'  => '12345',
                                       'productId'=> 'BATMAN')));

# Create a Binary serializer..other options
$serializer = new TBinarySerializer();
$msg_body   = $serializer->serialize($order);


$queue = 'orders.fullfilment';

$conn = new AMQPConnection(HOST, PORT, USER, PASS, VHOST);
$ch = $conn->channel();
$ch->queue_declare($queue, false, true, false, false);

for ($order_number=0;$order_number<6; $order_number++){
	$msg = new AMQPMessage($msg_body, array('content_type' => 'text/plain', 'delivery_mode' => 2));
	$ch->basic_publish($msg, '', $queue);
}
$ch->close();
$conn->close();
