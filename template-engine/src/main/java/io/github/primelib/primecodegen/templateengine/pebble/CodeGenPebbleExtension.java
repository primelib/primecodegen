package io.github.primelib.primecodegen.templateengine.pebble;

import io.github.primelib.primecodegen.templateengine.pebble.filter.LinePrefixFilter;
import io.github.primelib.primecodegen.templateengine.pebble.filter.PadRightFilter;
import io.github.primelib.primecodegen.templateengine.pebble.filter.WrapInFilter;
import io.github.primelib.primecodegen.templateengine.pebble.function.GetOrDefaultFunction;
import io.github.primelib.primecodegen.templateengine.pebble.function.NewLineFunction;
import io.github.primelib.primecodegen.templateengine.pebble.operator.StartsWithOperator;
import io.pebbletemplates.pebble.attributes.AttributeResolver;
import io.pebbletemplates.pebble.extension.Extension;
import io.pebbletemplates.pebble.extension.Filter;
import io.pebbletemplates.pebble.extension.Function;
import io.pebbletemplates.pebble.extension.NodeVisitorFactory;
import io.pebbletemplates.pebble.extension.Test;
import io.pebbletemplates.pebble.operator.BinaryOperator;
import io.pebbletemplates.pebble.operator.UnaryOperator;
import io.pebbletemplates.pebble.tokenParser.TokenParser;

import java.util.List;
import java.util.Map;

public class CodeGenPebbleExtension implements Extension {

    @Override
    public Map<String, Filter> getFilters() {
        return Map.of(
                "padright", new PadRightFilter(),
                "wrapin", new WrapInFilter(),
                "lineprefix", new LinePrefixFilter()
        );
    }

    @Override
    public Map<String, Test> getTests() {
        return null;
    }

    @Override
    public Map<String, Function> getFunctions() {
        return Map.of(
                "newline", new NewLineFunction(),
                "getOrDefault", new GetOrDefaultFunction()
        );
    }

    @Override
    public List<TokenParser> getTokenParsers() {
        return null;
    }

    @Override
    public List<BinaryOperator> getBinaryOperators() {
        return List.of(new StartsWithOperator());
    }

    @Override
    public List<UnaryOperator> getUnaryOperators() {
        return null;
    }

    @Override
    public Map<String, Object> getGlobalVariables() {
        return null;
    }

    @Override
    public List<NodeVisitorFactory> getNodeVisitors() {
        return null;
    }

    @Override
    public List<AttributeResolver> getAttributeResolver() {
        return null;
    }
}