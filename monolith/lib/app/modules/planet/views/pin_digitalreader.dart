import 'package:flutter/material.dart';
import 'package:monolith/pb/empirefox/firmata/config.pb.dart';
import 'package:monolith/pb/empirefox/firmata/instance.pb.dart';

class PinDigitalReader extends StatelessWidget {
  final Group_Pin groupPin;
  final Instance_Pin pin;
  final bool enabled;
  const PinDigitalReader(
      {Key? key,
      required this.groupPin,
      required this.pin,
      required this.enabled})
      : super(key: key);

  @override
  Widget build(BuildContext context) {
    late Icon data;
    if (!enabled) {
      data = const Icon(Icons.refresh, color: Colors.grey);
    } else {
      final type = groupPin.digitalReader;
      final triggered = pin.value == (type.lowLevelTrigger ? 0 : 1);
      late IconData icon;
      late Color color;
      if (type.alarm) {
        if (triggered) {
          icon = Icons.warning;
          color = Colors.red;
        } else {
          icon = Icons.check_circle;
          color = Colors.green;
        }
      } else {
        if (triggered) {
          icon = Icons.radio_button_checked;
          color = Colors.green;
        } else {
          icon = Icons.radio_button_unchecked;
          color = Colors.grey;
        }
      }
      data = Icon(icon, color: color);
    }
    return ListTile(
      title: Text(groupPin.nick),
      subtitle: Text(groupPin.desc),
      trailing: data,
      enabled: enabled,
    );
  }
}
