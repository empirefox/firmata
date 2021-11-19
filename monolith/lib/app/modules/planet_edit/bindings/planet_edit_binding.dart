import 'package:get/get.dart';

import '../controllers/planet_edit_controller.dart';

class PlanetEditBinding extends Bindings {
  @override
  void dependencies() {
    Get.lazyPut<PlanetEditController>(
      () => PlanetEditController(),
    );
  }
}
