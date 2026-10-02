import 'package:flutter_test/flutter_test.dart';
import 'package:presensigo_app/main.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  testWidgets('App renders login screen', (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues({});
    await tester.pumpWidget(const PresensiGoApp());
    await tester.pump(const Duration(milliseconds: 600));
    await tester.pump();
    expect(find.text('PresensiGo'), findsOneWidget);
  });
}
