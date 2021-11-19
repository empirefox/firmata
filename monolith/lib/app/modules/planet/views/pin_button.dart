import 'package:get/get.dart' show ExtensionSnackbar, Get;
import 'package:flutter/material.dart';
import 'package:monolith/pb/empirefox/firmata/config.pb.dart';

import '../types/types.dart';

class PinButton extends StatefulWidget {
  final Group_Pin groupPin;
  final TriggerCallback onTrigger;
  final bool enabled;
  const PinButton(
      {Key? key,
      required this.groupPin,
      required this.onTrigger,
      required this.enabled})
      : super(key: key);

  @override
  _PinButtonState createState() => _PinButtonState();
}

class _PinButtonState extends State<PinButton> {
  int _count = 0;
  Group_Pin get groupPin => widget.groupPin;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      title: Text(groupPin.nick),
      subtitle: Text(groupPin.desc),
      trailing: CircleAvatar(
        backgroundColor: widget.enabled ? Colors.blue : Colors.grey,
        child: Text('$_count'),
      ),
      enabled: widget.enabled,
      onTap: () {
        final ms = groupPin.button.triggerMs;
        if (ms == 0) {
          Get.snackbar('${groupPin.nick} operation', 'trigger time required');
          return;
        }

        widget.onTrigger.call(ms).then(
              (_) => setState(() => _count++),
              onError: (err) =>
                  Get.snackbar('${groupPin.nick} operation', 'error: $err'),
            );
      },
    );
  }
}
