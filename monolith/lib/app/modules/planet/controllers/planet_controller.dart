import 'dart:async';

import 'package:flutter/widgets.dart';
import 'package:get/get.dart' show Get, GetNavigation, GetxController, Inst;
import 'package:group_list_view/group_list_view.dart';
import 'package:monolith/app/data/model/model.dart';
import 'package:monolith/app/data/providers/transport.dart';
import 'package:monolith/app/data/services/storage.service.dart';
import 'package:monolith/app/data/services/transport.service.dart';
import 'package:monolith/app/routes/app_pages.dart';
import 'package:monolith/app/share/types.dart';
import 'package:monolith/pb/empirefox/firmata/config.pb.dart';
import 'package:monolith/pb/empirefox/firmata/instance.pb.dart';

import '../types/types.dart';

class PlanetController extends GetxController with WidgetsBindingObserver {
  final _ss = Get.find<StorageService>();
  final _ts = Get.find<TransportService>();

  Future<Transport>? _future;
  bool _futureDone = false;

  // requires future
  late final PlanetConfig config;
  // requires future
  late final Transport transport;
  List<Instance?> get instances => transport.instances;

  bool _appShown = true;
  StreamSubscription? _sub;

  @override
  void onInit() {
    super.onInit();
    WidgetsBinding.instance!.addObserver(this);
  }

  @override
  void onReady() {
    super.onReady();
  }

  @override
  void onClose() {
    _cancelListen();
    WidgetsBinding.instance!.removeObserver(this);
    super.onClose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    print('state changed');
    switch (state) {
      case AppLifecycleState.resumed:
        _appShown = true;
        break;
      case AppLifecycleState.inactive:
        _appShown = false;
        break;
      case AppLifecycleState.paused:
        _appShown = false;
        break;
      case AppLifecycleState.detached:
        _appShown = false;
        break;
    }
    _checkListenOrCancel();
  }

  // future must not be null to proccess.
  Future<Transport>? get future {
    if (_future == null && !_futureDone) {
      _futureDone = true;

      PlanetConfig? config = _ss.getxPlanet();
      if (config == null) {
        return null;
      }

      this.config = config;

      _future = _ts.getOrCreate(config).then((t) {
        transport = t;
        _checkListenOrCancel();
        return t;
      });
    }
    return _future;
  }

  void _checkListenOrCancel() {
    if (_appShown) {
      _listen();
    } else {
      _cancelListen();
    }
  }

  void _listen() {
    if (_sub == null && _future != null) {
      _sub = transport.onServerMessage.listen((_) => update());
      update();
    }
  }

  void _cancelListen() {
    _sub?.cancel();
    _sub = null;
  }

  //
  // View
  //

  void about() => Get.toNamed(Routes.ABOUT);

  bool isInstanceAlive(IndexPath index) => instances[index.section] != null;

  Group_Pin groupPin(IndexPath index) =>
      transport.groupVisiblePins[index.section][index.index];

  //
  // GroupListView
  //

  int countOfItemInSection(int section) =>
      transport.groupVisiblePins[section].length;

  //
  // GetBuilderFilters
  //

  GetBuilderFilter<PlanetController> detectOrAliveFilter(
          IndexPath index, Instance_Pin pin, Instance_Pin? detect) =>
      detect != null ? detectFilter(index, pin, detect) : aliveFilter(index);

  GetBuilderFilter<PlanetController> aliveFilter(IndexPath index) =>
      (_) => instances[index.section]?.firmataIndex ?? -1;

  GetBuilderFilter<PlanetController> valueFilter(
          IndexPath index, Instance_Pin pin) =>
      (_) => instances[index.section] == null ? double.nan : pin.value;

  GetBuilderFilter<PlanetController> detectFilter(
          IndexPath index, Instance_Pin pin, Instance_Pin? detect) =>
      (_) => instances[index.section] == null
          ? double.nan
          : (detect?.value ?? pin.value);

  TriggerCallback onTriggerButton(IndexPath index) {
    return (int ms) => transport.triggerDigitalPin(
          group: index.section,
          gpin: index.index,
          realtimeTriggerMs: ms,
        );
  }

  //
  // actions
  //

  TriggerCallback onTriggerSwitch(IndexPath index, Group_Pin groupPin) {
    return (int ms) {
      if (ms == 0) {
        return transport.setPinValue(
          group: index.section,
          gpin: index.index,
          value: transport
                  .instances[groupPin.firmataIndex]!.pins[groupPin.dx].value ^
              1,
        );
      }
      return transport.triggerDigitalPin(
        group: index.section,
        gpin: index.index,
        realtimeTriggerMs: ms,
      );
    };
  }

  TriggerCallback onTriggerNumberWriter(IndexPath index) {
    return (int value) => transport.setPinValue(
          group: index.section,
          gpin: index.index,
          value: value,
        );
  }
}
