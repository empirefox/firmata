import 'package:flutter/material.dart';
import 'package:monolith/pb/empirefox/firmata/config.pb.dart';
import 'package:monolith/pb/empirefox/firmata/instance.pb.dart';

class PinNumberReader extends StatelessWidget {
  final Group_Pin groupPin;
  final Instance_Pin pin;
  final bool enabled;
  const PinNumberReader(
      {Key? key,
      required this.groupPin,
      required this.pin,
      required this.enabled})
      : super(key: key);

  @override
  Widget build(BuildContext context) {
    return ListTile(
      title: Text(groupPin.nick),
      subtitle: Text(groupPin.desc),
      trailing: Text(
        '${pin.value}',
        style: TextStyle(color: enabled ? Colors.blue : Colors.grey),
      ),
      enabled: enabled,
    );
  }
}
