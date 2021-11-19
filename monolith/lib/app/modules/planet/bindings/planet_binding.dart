import 'package:get/get.dart';

import '../controllers/planet_controller.dart';

class PlanetBinding extends Bindings {
  @override
  void dependencies() {
    Get.lazyPut<PlanetController>(
      () => PlanetController(),
    );
  }
}
