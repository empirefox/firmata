import 'package:flutter/material.dart';
import 'package:get/get.dart' show ExtensionSnackbar, Get;
import 'package:monolith/app/data/providers/transport.dart';
import 'package:monolith/pb/empirefox/firmata/config.pb.dart';
import 'package:monolith/pb/empirefox/firmata/instance.pb.dart';

import '../types/types.dart';

class PinSwitch extends StatefulWidget {
  final Transport transport;
  final Group_Pin groupPin;
  final Instance_Pin pin;
  final Instance_Pin? detect;
  final TriggerCallback onTrigger;
  final bool enabled;
  const PinSwitch(
      {Key? key,
      required this.transport,
      required this.groupPin,
      required this.pin,
      required this.detect,
      required this.onTrigger,
      required this.enabled})
      : super(key: key);

  @override
  _PinSwitchState createState() => _PinSwitchState();
}

class _PinSwitchState extends State<PinSwitch> {
  Group_Pin get groupPin => widget.groupPin;
  Instance_Pin get pin => widget.pin;
  Instance_Pin? get detect => widget.detect;

  @override
  Widget build(BuildContext context) {
    final type = groupPin.switch_21;
    return SwitchListTile(
      title: Text(groupPin.nick),
      subtitle: Text(groupPin.desc),
      value: (detect?.value ?? pin.value) == (type.lowLevelTrigger ? 0 : 1),
      onChanged: widget.enabled
          ? (newValue) {
              widget.onTrigger.call(type.triggerMs).then(
                (_) {
                  if (detect == null) setState(() => pin.value ^= 1);
                },
                onError: (err) =>
                    Get.snackbar('${groupPin.nick} error', '$err'),
              );
            }
          : null,
    );
  }
}
