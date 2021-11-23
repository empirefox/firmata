import 'dart:developer';

import 'package:flutter/material.dart';
import 'package:flutter_easyloading/flutter_easyloading.dart' show EasyLoading;
import 'package:get/get.dart' show Get, GetBuilder, GetNavigation, Inst;
import 'package:group_list_view/group_list_view.dart';
import 'package:monolith/app/data/providers/transport.dart';
import 'package:monolith/app/routes/app_pages.dart';
import 'package:monolith/pb/empirefox/firmata/config.pb.dart';
import 'package:nil/nil.dart';

import '../controllers/planet_controller.dart';
import 'pin_button.dart';
import 'pin_digitalreader.dart';
import 'pin_error.dart';
import 'pin_numberreader.dart';
import 'pin_numberwriter.dart';
import 'pin_switch.dart';

class PlanetView extends StatelessWidget {
  final controller = Get.find<PlanetController>();
  final empty = Container();

  @override
  Widget build(BuildContext context) {
    final controller = Get.find<PlanetController>();
    final future = controller.future;
    if (future == null) {
      Get.offNamed(Routes.PLANETS);
      return empty;
    }
    return FutureBuilder(
      future: future,
      builder: (BuildContext context, AsyncSnapshot<Transport> snapshot) {
        final title = controller.config.viewName(snapshot.data?.config.nick);
        return Scaffold(
          appBar: AppBar(
            leading: IconButton(
              onPressed: Get.back,
              icon: Icon(Icons.chevron_left),
            ),
            title: Text(title),
            actions: [
              IconButton(
                onPressed: controller.about,
                icon: Icon(Icons.favorite),
              ),
            ],
          ),
          body: _futureBuilder(context, snapshot),
        );
      },
    );
  }

  Widget _futureBuilder(
      BuildContext context, AsyncSnapshot<Transport> snapshot) {
    switch (snapshot.connectionState) {
      case ConnectionState.none:
        return nil;
      case ConnectionState.waiting:
        EasyLoading.show(status: 'connecting...');
        return nil;
      case ConnectionState.active:
        return nil;
      case ConnectionState.done:
        EasyLoading.dismiss();
        if (snapshot.data == null) return nil;
        return _listView(context);
    }
  }

  Widget _listView(BuildContext context) {
    final transport = controller.transport;
    final groups = transport.config.groups;
    return GroupListView(
      sectionsCount: groups.length,
      countOfItemInSection: controller.countOfItemInSection,
      separatorBuilder: (context, index) => const SizedBox(height: 5),
      sectionSeparatorBuilder: (context, section) => const SizedBox(height: 10),
      itemBuilder: _itemBuilder,
      groupHeaderBuilder: _groupHeaderBuilder,
    );
  }

  Widget _groupHeaderBuilder(BuildContext context, int section) {
    final transport = controller.transport;
    final groups = transport.config.groups;
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 15, vertical: 8),
      child: ListTile(
        title: Text(
          groups[section].name,
          style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w600),
        ),
        subtitle:
            groups[section].desc.isEmpty ? null : Text(groups[section].desc),
      ),
    );
  }

  Widget _itemBuilder(BuildContext context, IndexPath index) {
    final transport = controller.transport;
    final groupPin = controller.groupPin(index);
    final pin = groupPin.getPin(transport);
    if (pin == null) {
      return PinError(groupPin: groupPin);
    }
    switch (groupPin.whichType()) {
      case Group_Pin_Type.button:
        return GetBuilder<PlanetController>(
          init: controller,
          filter: controller.aliveFilter(index),
          builder: (_) {
            return PinButton(
              groupPin: groupPin,
              onTrigger: controller.onTriggerButton(index),
              enabled: controller.isInstanceAlive(index),
            );
          },
        );
      case Group_Pin_Type.switch_21:
        final detect = groupPin.getDetectPin(transport);
        return GetBuilder<PlanetController>(
          init: controller,
          filter: controller.detectOrAliveFilter(index, pin, detect),
          builder: (_) {
            return PinSwitch(
              transport: transport,
              groupPin: groupPin,
              pin: pin,
              detect: detect,
              onTrigger: controller.onTriggerSwitch(index, groupPin),
              enabled: controller.isInstanceAlive(index),
            );
          },
        );
      case Group_Pin_Type.numberWriter:
        return GetBuilder<PlanetController>(
          init: controller,
          filter: controller.aliveFilter(index),
          builder: (_) {
            return PinNumberWriter(
              groupPin: groupPin,
              pin: pin,
              onTrigger: controller.onTriggerNumberWriter(index),
              enabled: controller.isInstanceAlive(index),
            );
          },
        );
      case Group_Pin_Type.digitalReader:
        return GetBuilder<PlanetController>(
          init: controller,
          filter: controller.aliveFilter(index),
          builder: (_) {
            return PinDigitalReader(
              groupPin: groupPin,
              pin: pin,
              enabled: controller.isInstanceAlive(index),
            );
          },
        );
      case Group_Pin_Type.numberReader:
        return GetBuilder<PlanetController>(
          init: controller,
          filter: controller.aliveFilter(index),
          builder: (_) {
            return PinNumberReader(
              groupPin: groupPin,
              pin: pin,
              enabled: controller.isInstanceAlive(index),
            );
          },
        );
      case Group_Pin_Type.hide:
        log('${groupPin.nick} error: hidden pins should not be handled');
        return PinError(groupPin: groupPin);
      case Group_Pin_Type.notSet:
        log('${groupPin.nick} bugs: type must be set');
        return PinError(groupPin: groupPin);
    }
  }
}
